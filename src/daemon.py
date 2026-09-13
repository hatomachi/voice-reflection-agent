import argparse
import logging
import signal
import sys
import time
from pathlib import Path

from .config import Config
from .git_sync import git_sync_inbound
from .processor import process_queue

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)
logger = logging.getLogger("voice_reflection.daemon")

running = True

def signal_handler(signum, frame):
    global running
    logger.info(f"Received signal {signum}. Shutting down gracefully...")
    running = False

def run_cycle(config: Config) -> int:
    """
    Single cycle:
    1. Outbound git fetch & pull (if updates exist)
    2. Process queue files
    """
    logger.debug("Starting sync cycle...")
    try:
        git_sync_inbound(config.vault_path)
    except Exception as e:
        logger.error(f"Error during git sync: {e}")

    try:
        count = process_queue(config)
        return count
    except Exception as e:
        logger.error(f"Error during queue processing: {e}")
        return 0

def main():
    parser = argparse.ArgumentParser(description="Voice Reflection Agent Daemon")
    parser.add_argument("--once", action="store_true", help="Run a single cycle and exit")
    parser.add_argument("--interval", type=int, default=None, help="Polling interval in seconds")
    parser.add_argument("--debug", action="store_true", help="Enable debug logging")
    args = parser.parse_args()

    if args.debug:
        logging.getLogger().setLevel(logging.DEBUG)

    config = Config()
    if args.interval:
        config.poll_interval_sec = args.interval

    logger.info(f"Initialized Voice Reflection Agent.")
    logger.info(f"Personal Vault: {config.vault_path}")
    logger.info(f"Queue Directory: {config.queue_dir}")
    logger.info(f"Reflections Directory: {config.reflections_dir}")
    logger.info(f"AGY Binary: {config.agy_bin}")

    if args.once:
        logger.info("Executing single cycle (--once)...")
        run_cycle(config)
        logger.info("Single cycle completed.")
        return

    # Daemon mode
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    logger.info(f"Starting daemon loop (polling every {config.poll_interval_sec}s)...")
    while running:
        run_cycle(config)
        # Sleep with short increments to respond quickly to signals
        for _ in range(config.poll_interval_sec):
            if not running:
                break
            time.sleep(1)

    logger.info("Daemon stopped.")

if __name__ == "__main__":
    main()
