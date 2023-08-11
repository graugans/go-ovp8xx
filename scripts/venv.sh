# shellcheck shell=sh

venv_repo_root=$(git rev-parse --show-toplevel)

python3 -m venv "$venv_repo_root"/.venv-pre-commit
. "$venv_repo_root"/.venv-pre-commit/bin/activate
pip install -r "$venv_repo_root"/requirements.txt

pre-commit install