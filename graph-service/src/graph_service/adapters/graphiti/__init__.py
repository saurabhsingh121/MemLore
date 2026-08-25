"""Graphiti adapter — sole import boundary for graphiti_core."""

from graph_service.adapters.graphiti.adapter import (
    GraphitiKnowledgeGraph,
    GraphitiUnavailableError,
    assert_memlore_result_shape,
)

__all__ = [
    "GraphitiKnowledgeGraph",
    "GraphitiUnavailableError",
    "assert_memlore_result_shape",
]
