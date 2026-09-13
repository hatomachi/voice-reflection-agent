import logging
from datetime import datetime
from pathlib import Path

logger = logging.getLogger("voice_reflection.context_loader")

WEEKDAY_JA = ["月", "火", "水", "木", "金", "土", "日"]

def format_date_with_day(dt: datetime) -> str:
    weekday = WEEKDAY_JA[dt.weekday()]
    return f"{dt.strftime('%Y-%m-%d')} ({weekday})"

def load_role_definitions(vault_path: Path) -> str:
    role_file = vault_path / "00_役割定義.md"
    if not role_file.exists():
        logger.warning(f"Role definition file not found: {role_file}")
        return ""
    try:
        return role_file.read_text(encoding="utf-8").strip()
    except Exception as e:
        logger.error(f"Failed to read role definition file: {e}")
        return ""

def load_dashboard_context(vault_path: Path) -> str:
    dashboard_file = vault_path / "00_Dashboard.md"
    if not dashboard_file.exists():
        logger.warning(f"Dashboard file not found: {dashboard_file}")
        return ""
    try:
        content = dashboard_file.read_text(encoding="utf-8")
        # Extract the focus areas and strategy maps if possible, or return first 100 lines
        lines = content.splitlines()
        extracted = lines[:80]
        return "\n".join(extracted).strip()
    except Exception as e:
        logger.error(f"Failed to read dashboard file: {e}")
        return ""

def load_recent_reflections(vault_path: Path, max_notes: int = 2) -> str:
    reflections_dir = vault_path / "00_Inbox" / "Reflections"
    if not reflections_dir.exists():
        return "（過去の振り返りログはまだありません）"

    notes = sorted(reflections_dir.glob("????-??-??.md"), reverse=True)
    if not notes:
        return "（過去の振り返りログはまだありません）"

    selected_notes = notes[:max_notes]
    summaries = []
    for note in selected_notes:
        try:
            content = note.read_text(encoding="utf-8").strip()
            # Omit audio transcription section to keep context concise
            if "<details>" in content:
                content = content.split("<details>")[0].strip()
            summaries.append(f"### {note.stem}\n{content}")
        except Exception as e:
            logger.warning(f"Failed to read reflection note {note}: {e}")

    return "\n\n".join(summaries) if summaries else "（過去の振り返りログはまだありません）"

def build_prompt(template_path: Path, vault_path: Path, dt: datetime, raw_transcription: str) -> str:
    if not template_path.exists():
        raise FileNotFoundError(f"Template not found at {template_path}")

    template = template_path.read_text(encoding="utf-8")
    date_with_day = format_date_with_day(dt)
    role_defs = load_role_definitions(vault_path)
    dashboard = load_dashboard_context(vault_path)
    recent_refs = load_recent_reflections(vault_path)

    prompt = template.replace("{{DATE_WITH_DAY}}", date_with_day)
    prompt = prompt.replace("{{ROLE_DEFINITIONS}}", role_defs)
    prompt = prompt.replace("{{DASHBOARD_CONTEXT}}", dashboard)
    prompt = prompt.replace("{{RECENT_REFLECTIONS}}", recent_refs)
    prompt = prompt.replace("{{RAW_TRANSCRIPTION}}", raw_transcription)

    return prompt
