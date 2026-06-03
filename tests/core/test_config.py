import pytest

from src.core.config import Config


def test_config_loads(tmp_path):
    config_file = tmp_path / "config.yaml"
    config_file.write_text("logging:\n  level: DEBUG\n")

    config = Config.from_yaml(config_file)
    assert config.logging.level == "DEBUG"


def test_config_defaults(tmp_path):
    config_file = tmp_path / "config.yaml"
    config_file.write_text("logging: {}\n")

    config = Config.from_yaml(config_file)
    assert config.logging.level == "INFO"


def test_config_env_override(tmp_path, monkeypatch):
    config_file = tmp_path / "config.yaml"
    config_file.write_text("logging:\n  level: DEBUG\n")

    monkeypatch.setenv("APP_LOGGING__LEVEL", "WARNING")
    config = Config.from_yaml(config_file)
    assert config.logging.level == "WARNING"


def test_config_missing_file():
    with pytest.raises(FileNotFoundError):
        Config.from_yaml("nonexistent.yaml")
