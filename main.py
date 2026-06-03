import asyncio

from src.core.config import Config, LoggingConfig
from src.core.logger import AppLogger


async def main() -> None:
    AppLogger(LoggingConfig())
    config = Config.from_yaml()
    log = AppLogger.get(__name__).bind(service="app")
    log.info("Config loaded", log_level=config.logging.level)

    # Code ...
    log.info("Starting application")


if __name__ == "__main__":
    asyncio.run(main())
