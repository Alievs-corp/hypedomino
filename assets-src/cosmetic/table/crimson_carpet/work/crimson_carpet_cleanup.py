import argparse
import os

import numpy as np
from PIL import Image
from scipy import ndimage

CANVAS = (1920, 1080)
SUPERSAMPLE = 2
MARGIN_TOLERANCE = 30
MARGIN_SHARE = 0.98
STRETCH_LIMIT = 1.1
RIM_WIDTH = 40
RIM_RADIUS = 28
INLAY_INSET = 54
INLAY_WIDTH = 4
INLAY_CORNER = 11
WOOD = {"hi": "#5A2414", "default": "#3A150B", "lo": "#240C06"}
GOLD = {"hi": "#FFF4C2", "default": "#E6C068", "mid": "#C99A3C", "lo": "#7A5418"}
CALM_RADIUS = (0.44, 0.40)
CALM_EDGE = 0.25
CALM_SOFTEN = 0.3
CALM_GAIN = 0.95
CALM_SIGMA = 12


def rgb(hex_colour):
    return np.array([int(hex_colour[i:i + 2], 16) for i in (1, 3, 5)], dtype=np.float64)


def margin_trim(pixels):
    ring = np.concatenate([pixels[:2].reshape(-1, 3), pixels[-2:].reshape(-1, 3), pixels[:, :2].reshape(-1, 3), pixels[:, -2:].reshape(-1, 3)])
    near = np.linalg.norm(pixels - np.median(ring, axis=0), axis=2) < MARGIN_TOLERANCE
    rows = np.where(near.mean(axis=1) < MARGIN_SHARE)[0]
    columns = np.where(near.mean(axis=0) < MARGIN_SHARE)[0]
    return columns[0], rows[0], columns[-1] + 1, rows[-1] + 1


def fit_to_canvas(image):
    width, height = image.size
    target = CANVAS[0] / CANVAS[1]
    stretch = min(max(target / (width / height), 1 / STRETCH_LIMIT), STRETCH_LIMIT)
    stretched_width = width * stretch
    if stretched_width / height > target:
        crop_width = height * target / stretch
        left = (width - crop_width) / 2
        box = (left, 0, left + crop_width, height)
    else:
        crop_height = stretched_width / target
        upper = (height - crop_height) / 2
        box = (0, upper, width, upper + crop_height)
    return image.resize(CANVAS, Image.LANCZOS, box=box), stretch, tuple(round(value, 1) for value in box)


def calm_centre(image):
    pixels = np.asarray(image).astype(np.float64)
    height, width = pixels.shape[:2]
    ys, xs = np.mgrid[0:height, 0:width] + 0.5
    dx = (xs - width / 2) / (width * CALM_RADIUS[0])
    dy = (ys - height / 2) / (height * CALM_RADIUS[1])
    weight = (np.clip(1 - np.hypot(dx, dy), 0, CALM_EDGE) / CALM_EDGE)[..., None]
    base = ndimage.gaussian_filter(pixels, (CALM_SIGMA, CALM_SIGMA, 0))
    calmed = base + (pixels - base) * (1 - CALM_SOFTEN * weight)
    toned = calmed * (1 - (1 - CALM_GAIN) * weight)
    return Image.fromarray(toned.round().clip(0, 255).astype(np.uint8), mode="RGB")


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


def front_layer(size):
    scale = SUPERSAMPLE
    big = (size[0] * scale, size[1] * scale)
    opening = rounded_rectangle(big, RIM_WIDTH * scale, RIM_RADIUS * scale)
    rim = ~opening
    light = light_from_above(rim, RIM_WIDTH * scale)
    colour = shade(light, {-1: rgb(WOOD["lo"]), 0: rgb(WOOD["default"]), 1: rgb(WOOD["hi"])})
    return downsample(colour, rim, size)


def inlay_mask(size, inset, width, corner):
    outer = rounded_rectangle(size, inset, 0.5)
    inner = rounded_rectangle(size, inset + width, 0.5)
    line = outer & ~inner
    ys, xs = np.mgrid[0:size[1], 0:size[0]] + 0.5
    centre = inset + width / 2
    for corner_x in (centre, size[0] - centre):
        for corner_y in (centre, size[1] - centre):
            line |= np.abs(xs - corner_x) + np.abs(ys - corner_y) <= corner
    return line


def middle_layer(size):
    scale = SUPERSAMPLE
    big = (size[0] * scale, size[1] * scale)
    line = inlay_mask(big, INLAY_INSET * scale, INLAY_WIDTH * scale, INLAY_CORNER * scale)
    light = light_from_above(line, INLAY_WIDTH * scale)
    colour = shade(light, {-1: rgb(GOLD["lo"]), -0.4: rgb(GOLD["mid"]), 0.2: rgb(GOLD["default"]), 1: rgb(GOLD["hi"])})
    return downsample(colour, line, size)


def downsample(colour, mask, size):
    alpha = mask.astype(np.float64) * 255
    premultiplied = np.dstack([colour * mask[..., None], alpha]).astype(np.float32)
    small = np.stack([np.asarray(Image.fromarray(premultiplied[..., channel], mode="F").resize(size, Image.BOX)) for channel in range(4)], axis=-1)
    opacity = np.maximum(small[..., 3:], 1e-6) / 255
    straight = np.where(small[..., 3:] > 0, small[..., :3] / opacity, 0)
    return Image.fromarray(np.dstack([straight, small[..., 3:]]).round().clip(0, 255).astype(np.uint8), mode="RGBA")


def main():
    parser = argparse.ArgumentParser(prog="table_cleanup.py")
    parser.add_argument("slug")
    parser.add_argument("back")
    parser.add_argument("work")
    parser.add_argument("export")
    arguments = parser.parse_args()
    concept = Image.open(arguments.back).convert("RGB")
    trim = margin_trim(np.asarray(concept).astype(np.float64))
    back, stretch, box = fit_to_canvas(concept.crop(trim))
    back = calm_centre(back).convert("RGBA")
    middle = middle_layer(CANVAS)
    front = front_layer(CANVAS)
    for name, layer in (("back", back), ("middle", middle), ("front", front)):
        layer.save(os.path.join(arguments.work, f"{arguments.slug}_{name}.png"))
    flattened = Image.alpha_composite(Image.alpha_composite(back, middle), front)
    flattened.save(os.path.join(arguments.export, "static.png"))
    print(f"{arguments.slug}: trim {tuple(int(value) for value in trim)}, stretch {stretch:.3f}, crop {box}, canvas {CANVAS[0]}x{CANVAS[1]}")


main()
