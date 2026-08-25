FROM python:3.12-slim

WORKDIR /app

RUN pip install --no-cache-dir uv

COPY pyproject.toml README.md ./
COPY src ./src

RUN uv sync --no-dev

ENV MEMLORE_GRAPH_SERVICE_HOST=0.0.0.0
ENV MEMLORE_GRAPH_SERVICE_PORT=8090

EXPOSE 8090

CMD ["uv", "run", "uvicorn", "graph_service.adapters.http.app:create_app", "--factory", "--host", "0.0.0.0", "--port", "8090"]
