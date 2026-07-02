# Python Certification Notes

## Status: Not started

## Known Issues

*None yet — certification pending.*

## Architecture Notes

- Python buildpack **creates a virtual environment** before installing dependencies.
  - Windows: `python -m venv venv` → `.\venv\Scripts\pip install -r requirements.txt`
  - Unix: `python3 -m venv venv` → `./venv/bin/pip install -r requirements.txt`
- All start commands use the venv Python interpreter (e.g., `.\venv\Scripts\python app.py`).
- Detection logic checks for `requirements.txt`, `setup.py`, `setup.cfg`, or `Pipfile`.
- Framework detection:
  - `manage.py` → Django (`manage.py runserver`)
  - `wsgi.py` → Gunicorn or Uvicorn (detected from requirements.txt)
  - `app.py` → simple Python script
  - `main.py` → simple Python script
- `PYTHONUNBUFFERED=1` and `PYTHONDONTWRITEBYTECODE=1` are set.

## Potential Issues

- `pip install` may fail if system Python doesn't have ensurepip (common on minimal Docker images).
- Django requires `settings.py` configuration — may need `DJANGO_SETTINGS_MODULE` env var.
- Gunicorn on Windows requires `waitress` as an alternative.
