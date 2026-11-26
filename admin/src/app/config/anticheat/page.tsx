import { KeyedConfigPage } from "../KeyedConfigPage";
import { publishAntiCheat, rollbackAntiCheat } from "./actions";

export const dynamic = "force-dynamic";

export default function AntiCheatPage() {
  return (
    <KeyedConfigPage
      path="/admin/config/anticheat"
      object="anti_cheat_config"
      text={{ title: "anticheat.title", loadFailed: "anticheat.load_failed", intro: "anticheat.intro", historyEmpty: "anticheat.history_empty" }}
      publish={publishAntiCheat}
      rollback={rollbackAntiCheat}
    />
  );
}
