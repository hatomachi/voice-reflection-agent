import json
import logging
import shutil
from datetime import datetime
from pathlib import Path
from typing import Optional

from .config import Config
from .context_loader import build_prompt
from .agy_runner import run_agy_reflection
from .git_sync import git_commit_and_push

logger = logging.getLogger("voice_reflection.processor")

def parse_queue_file(file_path: Path) -> tuple[datetime, str]:
    """
    Parses a queue file (JSON or TXT) and returns (datetime, raw_transcription).
    """
    now = datetime.now()
    if file_path.suffix.lower() == ".json":
        data = json.loads(file_path.read_text(encoding="utf-8"))
        raw_text = data.get("text", "").strip()

        # Determine date
        date_str = data.get("date")
        timestamp_str = data.get("timestamp")
        dt = now
        if timestamp_str:
            try:
                dt = datetime.fromisoformat(timestamp_str)
            except Exception:
                pass
        elif date_str:
            try:
                dt = datetime.strptime(date_str, "%Y-%m-%d")
            except Exception:
                pass

        return dt, raw_text

    elif file_path.suffix.lower() in [".txt", ".md"]:
        raw_text = file_path.read_text(encoding="utf-8").strip()
        # Try parse timestamp from filename: YYYY-MM-DD-HHmmss.txt
        stem = file_path.stem
        try:
            dt = datetime.strptime(stem[:10], "%Y-%m-%d")
        except Exception:
            dt = now
        return dt, raw_text

    else:
        raise ValueError(f"Unsupported queue file format: {file_path}")

def save_reflection(reflections_dir: Path, dt: datetime, markdown_content: str) -> Path:
    reflections_dir.mkdir(parents=True, exist_ok=True)
    date_str = dt.strftime("%Y-%m-%d")
    output_file = reflections_dir / f"{date_str}.md"

    if output_file.exists():
        # Append as an update/additional reflection
        time_str = dt.strftime("%H:%M")
        existing = output_file.read_text(encoding="utf-8")
        updated = (
            f"{existing}\n\n---\n\n"
            f"## 🎙️ 追加セルフリフレクション ({time_str})\n\n"
            f"{markdown_content}\n"
        )
        output_file.write_text(updated, encoding="utf-8")
        logger.info(f"Appended additional reflection to {output_file}")
    else:
        output_file.write_text(markdown_content, encoding="utf-8")
        logger.info(f"Created new reflection at {output_file}")

    return output_file

def process_queue(config: Config) -> int:
    """
    Processes all pending queue files in config.queue_dir.
    Returns the number of successfully processed files.
    """
    if not config.queue_dir.exists():
        logger.debug(f"Queue dir does not exist: {config.queue_dir}")
        return 0

    # Look for .json and .txt files, ignore hidden files like .gitkeep
    queue_files = sorted([
        f for f in config.queue_dir.iterdir()
        if f.is_file() and not f.name.startswith(".") and f.suffix.lower() in [".json", ".txt", ".md"]
    ])

    if not queue_files:
        logger.debug("No pending queue files.")
        return 0

    logger.info(f"Found {len(queue_files)} pending queue file(s).")
    processed_count = 0
    modified_files = []

    for qf in queue_files:
        logger.info(f"Processing queue file: {qf.name}")
        try:
            dt, raw_text = parse_queue_file(qf)
            if not raw_text:
                logger.warning(f"Queue file {qf.name} is empty. Deleting...")
                qf.unlink()
                modified_files.append(str(qf.relative_to(config.vault_path)))
                continue

            prompt = build_prompt(config.template_path, config.vault_path, dt, raw_text)
            reflection_md = run_agy_reflection(prompt, config.agy_bin, timeout_sec=config.agy_timeout_sec)

            output_file = save_reflection(config.reflections_dir, dt, reflection_md)

            # Archive or delete queue file
            qf.unlink()

            modified_files.append(str(output_file.relative_to(config.vault_path)))
            modified_files.append(str(qf.relative_to(config.vault_path)))
            processed_count += 1

        except Exception as e:
            logger.error(f"Failed to process queue file {qf.name}: {e}", exc_info=True)

    if processed_count > 0:
        commit_msg = f"feat(reflection): process {processed_count} voice reflection(s)"
        git_commit_and_push(config.vault_path, [config.reflections_rel_path, config.queue_rel_path], commit_msg)

    return processed_count
