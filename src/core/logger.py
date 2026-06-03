import logging
import sys
from pathlib import Path

import structlog

from src.core.config import LoggingConfig

_shared_processors: list[structlog.types.Processor] = [
    structlog.contextvars.merge_contextvars,
    structlog.stdlib.add_log_level,
    structlog.stdlib.add_logger_name,
    structlog.stdlib.PositionalArgumentsFormatter(),
    structlog.processors.TimeStamper(fmt="iso"),
    structlog.processors.StackInfoRenderer(),
]


class AppLogger:
    def __init__(self, config: LoggingConfig, log_dir: str = "logs") -> None:
        Path(log_dir).mkdir(parents=True, exist_ok=True)
        level = getattr(logging, config.level.upper(), logging.INFO)

        structlog.configure(
            processors=_shared_processors + [structlog.stdlib.ProcessorFormatter.wrap_for_formatter],
            logger_factory=structlog.stdlib.LoggerFactory(),
            wrapper_class=structlog.stdlib.BoundLogger,
            cache_logger_on_first_use=True,
        )

        file_formatter = structlog.stdlib.ProcessorFormatter(
            foreign_pre_chain=_shared_processors,
            processors=[
                structlog.stdlib.ProcessorFormatter.remove_processors_meta,
                structlog.processors.JSONRenderer(),
            ],
        )

        console_formatter = structlog.stdlib.ProcessorFormatter(
            foreign_pre_chain=_shared_processors,
            processors=[
                structlog.stdlib.ProcessorFormatter.remove_processors_meta,
                structlog.dev.ConsoleRenderer(colors=True),
            ],
        )

        file_handler = logging.FileHandler(Path(log_dir) / "app.log", encoding="utf-8")
        file_handler.setFormatter(file_formatter)

        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setFormatter(console_formatter)

        root = logging.getLogger()
        root.addHandler(file_handler)
        root.addHandler(console_handler)
        root.setLevel(level)

    @staticmethod
    def get(name: str) -> structlog.stdlib.BoundLogger:
        return structlog.get_logger(name)
