import argparse
import json
import os

import numpy as np
from PIL import Image
from scipy import ndimage

CANVAS = (1920, 1080)
KNOT = 4
SEED = 169
SUPERSAMPLE = 2
RIM_WIDTH = 40
RIM_RADIUS = 28
INLAY_INSET = 54
INLAY_WIDTH = 4
INLAY_CORNER = 11
REFERENCE = CANVAS
EDGE_DEPTH = 168
EDGE_FADE = 16
EDGE_LIGHT = 120
EDGE_STEPS = range(150, 486, 6)
EDGE_SHORTLIST = 40
CORNER_MARGIN = 30
SQUARE_REACH = 520
PALETTE = {
    "field": (118, 22, 30),
    "fieldDeep": (98, 16, 25),
    "fieldMotif": (132, 34, 36),
    "medallion": (126, 27, 33),
    "ochre": (156, 104, 50),
    "ivory": (212, 192, 154),
    "indigo": (32, 36, 78),
    "indigoDeep": (21, 23, 52),
}
BAND = (("indigoDeep", 0, 10), ("field", 10, 16), ("ivory", 16, 18), ("indigo", 18, 42), ("ivory", 42, 44), ("field", 44, 47))
FIELD_PERIOD = 30
BORDER_PERIOD = 24
MEDALLION = (150, 84, 4)
VIGNETTE = 0.3
JITTER = 0.035
ABRASH = 0.03
KNOT_EDGE = 0.9


def colour(name):
    return np.array(PALETTE[name], dtype=np.float64)


def token(tokens, path):
    value = tokens["color"]
    for key in path.split("."):
        value = value[key]
    return np.array([int(value[i:i + 2], 16) for i in (1, 3, 5)], dtype=np.float64)


def knot_grid():
    columns, rows = CANVAS[0] // KNOT, CANVAS[1] // KNOT
    v, u = np.mgrid[0:rows, 0:columns]
    return u, v, columns, rows


def border_coordinates(u, v, columns, rows, middle):
    distances = np.stack([v, rows - 1 - v, u, columns - 1 - u])
    nearest = distances.argmin(axis=0)
    across = distances.min(axis=0)
    run = np.where(nearest < 2, columns - 1 - 2 * middle, rows - 1 - 2 * middle)
    position = np.where(nearest < 2, u, v) - middle
    count = np.maximum(2, np.round(run / BORDER_PERIOD / 2) * 2)
    period = run / count
    index = np.round(position / period)
    return across, position - index * period, index


def flower(dx, dy, size):
    centre = np.abs(dx) + np.abs(dy) <= size * 0.34
    petals = np.zeros_like(centre)
    for ox, oy in ((1, 0), (-1, 0), (0, 1), (0, -1)):
        petals |= np.abs(dx - ox * size * 0.62) + np.abs(dy - oy * size * 0.62) <= size * 0.3
    return centre, petals


def stepped_diamond(dx, dy, size):
    reach = np.abs(dx) + np.abs(dy)
    outline = (reach <= size) & (reach > size - 2)
    core = reach <= size * 0.42
    return outline, core


def field(u, v, columns, rows, image):
    cu, cv = u - columns / 2 + 0.5, v - rows / 2 + 0.5
    lu = np.mod(cu, FIELD_PERIOD) - FIELD_PERIOD / 2
    lv = np.mod(cv, FIELD_PERIOD) - FIELD_PERIOD / 2
    lattice = np.abs(np.abs(lu) + np.abs(lv) - FIELD_PERIOD / 2) < 0.75
    image[lattice] = colour("fieldDeep")
    centre, petals = flower(lu, lv, 6)
    image[petals] = colour("fieldMotif")
    image[centre] = colour("fieldDeep")
    reach = np.abs(cu) / MEDALLION[0] + np.abs(cv) / MEDALLION[1]
    image[reach < 1] = colour("medallion")
    image[np.abs(reach - 1) * MEDALLION[1] < 1] = colour("fieldDeep")
    image[np.abs(reach - 1 - MEDALLION[2] / MEDALLION[1]) * MEDALLION[1] < 0.6] = colour("fieldMotif")


def border(u, v, columns, rows, image):
    band = BAND[3]
    middle = (band[1] + band[2] - 1) / 2
    across, dx, index = border_coordinates(u, v, columns, rows, middle)
    for name, start, end in BAND:
        image[(across >= start) & (across < end)] = colour(name)
    dy = across - middle
    inside = (across >= band[1]) & (across < band[2])
    diamond = np.mod(index, 2) == 0
    outline, core = stepped_diamond(dx, dy, 10)
    centre, petals = flower(dx, dy, 8)
    image[inside & diamond & outline] = colour("ochre")
    image[inside & diamond & core] = colour("field")
    image[inside & ~diamond & petals] = colour("ivory")
    image[inside & ~diamond & centre] = colour("ochre")
    guard = BAND[1]
    dots = (across >= guard[1]) & (across < guard[2]) & (np.abs(np.mod(dx, 6) - 3) + np.abs(across - (guard[1] + guard[2] - 1) / 2) <= 1)
    image[dots] = colour("indigoDeep")


def wool(knots, rows, rng):
    jitter = 1 + rng.uniform(-JITTER, JITTER, knots.shape[:2])
    abrash = 1 + ABRASH * np.sin(np.arange(rows) / rows * np.pi * 5 + rng.uniform(0, np.pi))
    shaded = knots * (jitter * abrash[:, None])[..., None]
    pixels = np.repeat(np.repeat(shaded, KNOT, axis=0), KNOT, axis=1)
    edge = np.ones(KNOT)
    edge[0] = edge[-1] = KNOT_EDGE
    pixels *= np.outer(np.tile(edge, rows), np.tile(edge, knots.shape[1]))[..., None]
    pixels = ndimage.gaussian_filter(pixels, (0.6, 0.6, 0))
    fibre = 1 + rng.normal(0, 0.02, pixels.shape[:2])
    return pixels * fibre[..., None]


def lamp(size):
    ys, xs = np.mgrid[0:size[1], 0:size[0]] + 0.5
    dx = (xs - size[0] / 2) / (size[0] / 2)
    dy = (ys - size[1] * 0.46) / (size[1] / 2)
    falloff = np.clip(np.hypot(dx * 0.8, dy) / np.hypot(0.8, 1.08), 0, 1)
    return 1 - VIGNETTE * falloff ** 1.6


def back_layer():
    rng = np.random.default_rng(SEED)
    u, v, columns, rows = knot_grid()
    knots = np.empty((rows, columns, 3))
    knots[:] = colour("field")
    field(u, v, columns, rows, knots)
    border(u, v, columns, rows, knots)
    pixels = wool(knots, rows, rng) * lamp(CANVAS)[..., None]
    pixels, step, start, score = periodic_edges(pixels, EDGE_DEPTH)
    opaque = np.full(pixels.shape[:2], 255.0)
    return Image.fromarray(np.dstack([pixels, opaque]).round().clip(0, 255).astype(np.uint8), mode="RGBA"), step, start, score


def edge_light(band):
    low = ndimage.gaussian_filter1d(band, EDGE_LIGHT, axis=1, mode="reflect") + 1
    return band / low, (low + low[:, ::-1]) / 2


def seam_cost(detail, step, start, reach):
    return np.abs(detail[:, start + step - reach:start + step + reach] - detail[:, start - reach:start + reach]).mean()


def mirror_cost(detail, start, reach=24):
    return np.abs(detail[:, start - reach:start][:, ::-1] - detail[:, start:start + reach]).mean()


def repeat_band(detail, step, start):
    width = detail.shape[1]
    end = width - start
    xs = np.arange(width)
    inside = (xs >= start) & (xs < end)
    result = detail[:, np.where(inside, start + (xs - start) % step, xs)]
    offsets = np.arange(-EDGE_FADE, EDGE_FADE)
    weight = ((offsets + EDGE_FADE + 0.5) / (2 * EDGE_FADE)).reshape((1, -1) + (1,) * (detail.ndim - 2))
    carried = detail[:, start + step + offsets]
    for seam in [*range(start + step, end, step), end]:
        arriving = detail[:, (start if seam < end else end) + offsets]
        result[:, seam + offsets] = carried * (1 - weight) + arriving * weight
    return result


def best_edges(details, depth, width):
    candidates = []
    for step in EDGE_STEPS:
        for count in range(2, width // step + 1):
            start, odd = divmod(width - count * step, 2)
            if odd or start < depth + CORNER_MARGIN or start > SQUARE_REACH - step // 4:
                continue
            cost = sum(seam_cost(detail, step, start, 12) + seam_cost(detail, step, start, 48) / 2 + mirror_cost(detail, start) / 2 for detail in details)
            candidates.append((cost, step, start))
    scored = []
    for cost, step, start in sorted(candidates)[:EDGE_SHORTLIST]:
        drift = sum(np.abs(repeat_band(detail, step, start) - detail)[:, start:width - start].mean() for detail in details)
        scored.append((cost + drift / 2, step, start))
    return min(scored)


def periodic_edges(pixels, depth):
    height, width = pixels.shape[:2]
    bands = (slice(0, depth), slice(height - depth, height))
    parts = [edge_light(pixels[rows]) for rows in bands]
    score, step, start = best_edges([detail.mean(axis=2) for detail, _ in parts], depth, width)
    inward = np.clip((depth - np.arange(depth) - 0.5) / EDGE_FADE, 0, 1)
    result = pixels.copy()
    for rows, (detail, light), weight in zip(bands, parts, (inward, inward[::-1])):
        rebuilt = repeat_band(detail, step, start) * light
        result[rows] = rebuilt * weight[:, None, None] + pixels[rows] * (1 - weight[:, None, None])
    return result, step, start, score


def write_slice(export, depth, step, size):
    across, down = size[0] / REFERENCE[0], size[1] / REFERENCE[1]
    values = {"depth": depth * down, "step": step * across}
    if any(value != int(value) for value in values.values()):
        raise ValueError(f"slice {values} is not whole pixels on the {size[0]}x{size[1]} canvas")
    os.makedirs(export, exist_ok=True)
    with open(os.path.join(export, "slice.json"), "w", encoding="utf-8", newline="\n") as target:
        json.dump({key: int(value) for key, value in values.items()}, target)
        target.write("\n")


def rounded_rectangle(size, inset, radius):
    width, height = size
    ys, xs = np.mgrid[0:height, 0:width] + 0.5
    inner_x = np.clip(xs, inset + radius, width - inset - radius)
    inner_y = np.clip(ys, inset + radius, height - inset - radius)
    return np.hypot(xs - inner_x, ys - inner_y) <= radius


def light_from_above(mask, width):
    depth = np.minimum(ndimage.distance_transform_edt(mask), width / 2) / (width / 2)
    height = np.sin(depth * np.pi / 2)
    slope = np.gradient(height, axis=0)
    return np.clip(slope * width / 2, -1, 1)


def shade(light, stops):
    keys = sorted(stops)
    positions = np.interp(light, keys, range(len(keys)))
    lower = np.floor(positions).astype(int).clip(0, len(keys) - 2)
    weight = (positions - lower)[..., None]
    colours = np.array([stops[key] for key in keys])
    return colours[lower] * (1 - weight) + colours[lower + 1] * weight


def downsample(colours, mask, size):
    alpha = mask.astype(np.float64) * 255
    premultiplied = np.dstack([colours * mask[..., None], alpha]).astype(np.float32)
    small = np.stack([np.asarray(Image.fromarray(premultiplied[..., channel], mode="F").resize(size, Image.BOX)) for channel in range(4)], axis=-1)
    opacity = np.maximum(small[..., 3:], 1e-6) / 255
    straight = np.where(small[..., 3:] > 0, small[..., :3] / opacity, 0)
    return Image.fromarray(np.dstack([straight, small[..., 3:]]).round().clip(0, 255).astype(np.uint8), mode="RGBA")


def front_layer(tokens):
    big = (CANVAS[0] * SUPERSAMPLE, CANVAS[1] * SUPERSAMPLE)
    rim = ~rounded_rectangle(big, RIM_WIDTH * SUPERSAMPLE, RIM_RADIUS * SUPERSAMPLE)
    light = light_from_above(rim, RIM_WIDTH * SUPERSAMPLE)
    stops = {-1: token(tokens, "wood.lo"), 0: token(tokens, "wood.default"), 1: token(tokens, "wood.hi")}
    return downsample(shade(light, stops), rim, CANVAS)


def inlay_mask(size, inset, width, corner):
    line = rounded_rectangle(size, inset, 0.5) & ~rounded_rectangle(size, inset + width, 0.5)
    ys, xs = np.mgrid[0:size[1], 0:size[0]] + 0.5
    centre = inset + width / 2
    for corner_x in (centre, size[0] - centre):
        for corner_y in (centre, size[1] - centre):
            line |= np.abs(xs - corner_x) + np.abs(ys - corner_y) <= corner
    return line


def middle_layer(tokens):
    big = (CANVAS[0] * SUPERSAMPLE, CANVAS[1] * SUPERSAMPLE)
    line = inlay_mask(big, INLAY_INSET * SUPERSAMPLE, INLAY_WIDTH * SUPERSAMPLE, INLAY_CORNER * SUPERSAMPLE)
    light = light_from_above(line, INLAY_WIDTH * SUPERSAMPLE)
    stops = {-1: token(tokens, "gold.lo"), -0.4: token(tokens, "gold.mid"), 0.2: token(tokens, "gold.default"), 1: token(tokens, "gold.hi")}
    return downsample(shade(light, stops), line, CANVAS)


def main():
    parser = argparse.ArgumentParser(prog="starter_carpet.py")
    parser.add_argument("tokens")
    parser.add_argument("work")
    parser.add_argument("export")
    arguments = parser.parse_args()
    with open(arguments.tokens, encoding="utf-8") as source:
        tokens = json.load(source)
    back, step, start, score = back_layer()
    layers = (("back", back), ("middle", middle_layer(tokens)), ("front", front_layer(tokens)))
    os.makedirs(arguments.work, exist_ok=True)
    os.makedirs(arguments.export, exist_ok=True)
    for name, layer in layers:
        layer.save(os.path.join(arguments.work, f"starter_{name}.png"), optimize=True)
    flattened = layers[0][1]
    for _, layer in layers[1:]:
        flattened = Image.alpha_composite(flattened, layer)
    flattened.save(os.path.join(arguments.export, "static.png"), optimize=True)
    write_slice(arguments.export, EDGE_DEPTH, step, CANVAS)
    print(f"starter: canvas {CANVAS[0]}x{CANVAS[1]}, knot {KNOT} px, seed {SEED}, edges depth {EDGE_DEPTH} step {step} start {start} score {score:.4f}")


main()
