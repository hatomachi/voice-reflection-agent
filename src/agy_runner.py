import subprocess
import logging
import re
from pathlib import Path

logger = logging.getLogger("voice_reflection.agy_runner")

ANSI_ESCAPE = re.compile(r'\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])')

def strip_ansi(text: str) -> str:
    return ANSI_ESCAPE.sub('', text)

def clean_markdown_output(raw_output: str) -> str:
    cleaned = strip_ansi(raw_output).strip()
    # If the response was wrapped in ```markdown ... ```, extract the content
    lines = cleaned.splitlines()
    if len(lines) >= 2:
        if lines[0].strip().startswith("```markdown") and lines[-1].strip() == "```":
            cleaned = "\n".join(lines[1:-1]).strip()
        elif lines[0].strip().startswith("```") and lines[-1].strip() == "```":
            cleaned = "\n".join(lines[1:-1]).strip()
    return cleaned

def run_agy_reflection(
    prompt: str,
    agy_bin: Path,
    timeout_sec: int = 300,
    max_retries: int = 1
) -> str:
    """
    Executes agy CLI in print mode to generate deep reflection.
    """
    if not agy_bin.exists():
        raise FileNotFoundError(f"agy binary not found at: {agy_bin}")

    cmd = [
        str(agy_bin),
        "-p",
        prompt,
        "--dangerously-skip-permissions"
    ]

    for attempt in range(1 + max_retries):
        try:
            logger.info(f"Running agy CLI (attempt {attempt + 1})...")
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=timeout_sec,
                check=False
            )
            if result.returncode == 0:
                output = clean_markdown_output(result.stdout)
                if output:
                    logger.info("Successfully received reflection from agy")
                    return output
                else:
                    logger.warning("agy returned empty output")
            else:
                logger.error(f"agy failed with code {result.returncode}: {result.stderr.strip()}")
        except subprocess.TimeoutExpired:
            logger.error(f"agy timed out after {timeout_sec}s (attempt {attempt + 1})")
        except Exception as e:
            logger.error(f"Unexpected error running agy: {e}")

    raise RuntimeError("Failed to generate reflection via agy after all retries")
