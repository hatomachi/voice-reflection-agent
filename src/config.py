import os
from pathlib import Path
from dataclasses import dataclass

@dataclass
class Config:
    vault_path: Path = Path(os.environ.get("PERSONAL_VAULT_PATH", "/Users/s-ikari/work/personal-vault"))
    agy_bin: Path = Path(os.environ.get("AGY_BIN_PATH", "/Users/s-ikari/.local/bin/agy"))
    template_path: Path = Path(__file__).resolve().parent.parent / "templates" / "reflection_prompt.md"
    queue_rel_path: str = "00_Inbox/queue"
    reflections_rel_path: str = "00_Inbox/Reflections"
    poll_interval_sec: int = int(os.environ.get("POLL_INTERVAL_SEC", "60"))
    agy_timeout_sec: int = int(os.environ.get("AGY_TIMEOUT_SEC", "300"))

    @property
    def queue_dir(self) -> Path:
        return self.vault_path / self.queue_rel_path

    @property
    def reflections_dir(self) -> Path:
        return self.vault_path / self.reflections_rel_path
