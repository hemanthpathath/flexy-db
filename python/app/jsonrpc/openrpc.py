"""
OpenRPC specification generator for JSON-RPC methods.

Auto-generates OpenRPC spec from registered JSON-RPC methods using introspection.
"""

import inspect
import json
from typing import Any, Dict, List, Optional

# OpenRPC spec metadata
OPENRPC_VERSION = "1.2.6"
SERVICE_NAME = "flex-db"
SERVICE_VERSION = "1.0.0"


def get_type_schema(param_type: type, default_value: Any = None) -> Dict[str, Any]:
    """Convert Python type to JSON Schema type."""
    type_mapping = {
        str: {"type": "string"},
        int: {"type": "integer"},
        float: {"type": "number"},
        bool: {"type": "boolean"},
        dict: {"type": "object"},
        list: {"type": "array"},
    }
    
    # Handle Optional types (Union[X, None])
    if hasattr(param_type, '__origin__'):
        origin = param_type.__origin__
        
        # Handle Optional[X] or Union[X, None]
        if origin is type(None) or (hasattr(origin, '__name__') and origin.__name__ == 'Union'):
            args = getattr(param_type, '__args__', [])
            # Filter out None type
            non_none_args = [arg for arg in args if arg is not type(None)]
            if non_none_args:
                schema = get_type_schema(non_none_args[0], default_value)
                if default_value is not None:
                    schema["default"] = default_value
                elif default_value == "":
                    schema["default"] = ""
                return schema
        
        # Handle Dict[str, Any] or Dict[str, Any] = None
        if origin is dict:
            return {"type": "object", "additionalProperties": True}
        
        # Handle List types
        if origin is list:
            return {"type": "array", "items": {"type": "object"}}
    
    # Direct type mapping
    if param_type in type_mapping:
        schema = type_mapping[param_type].copy()
        if default_value is not None:
            schema["default"] = default_value
        elif default_value == "":
            schema["default"] = ""
        return schema
    
    # Default to string if unknown
    schema = {"type": "string"}
    if default_value is not None:
        schema["default"] = default_value
    elif default_value == "":
        schema["default"] = ""
    return schema


def extract_method_info(func) -> Optional[Dict[str, Any]]:
    """Extract method information from a function for OpenRPC spec."""
    try:
        sig = inspect.signature(func)
        doc = inspect.getdoc(func) or ""
        
        params = []
        param_descriptions = {}
        
        # Parse docstring for parameter descriptions
        if doc:
            lines = doc.split('\n')
            for line in lines:
                line = line.strip()
                if ':' in line and not line.startswith('Returns'):
                    parts = line.split(':', 1)
                    if len(parts) == 2:
                        param_name = parts[0].strip()
                        param_descriptions[param_name] = parts[1].strip()
        
        # Extract parameters from signature
        for param_name, param in sig.parameters.items():
            if param_name == 'self':
                continue
            
            param_type = param.annotation if param.annotation != inspect.Parameter.empty else str
            default_value = param.default if param.default != inspect.Parameter.empty else None
            
            param_schema = get_type_schema(param_type, default_value)
            
            param_info = {
                "name": param_name,
                "schema": param_schema,
            }
            
            if param_name in param_descriptions:
                param_info["description"] = param_descriptions[param_name]
            elif default_value is not None and default_value != "":
                param_info["description"] = f"Optional. Default: {default_value}"
            elif default_value == "":
                param_info["description"] = "Optional. Default: empty string"
            
            if default_value is not None:
                param_info["required"] = False
            else:
                param_info["required"] = True
            
            params.append(param_info)
        
        # Get first line of docstring as description
        description = doc.split('\n')[0] if doc else f"{func.__name__} method"
        
        return {
            "name": func.__name__,
            "description": description,
            "params": params,
            "result": {
                "name": "result",
                "schema": {
                    "type": "object",
                    "description": "Method result object"
                }
            }
        }
    except Exception as e:
        # If we can't extract info, return None to skip this method
        return None


def generate_openrpc_spec() -> Dict[str, Any]:
    """Generate OpenRPC specification from registered methods."""
    methods_list = []
    
    # Import handlers module to inspect methods
    from app.jsonrpc import handlers
    
    # Get all async functions from handlers module
    # These are our JSON-RPC methods decorated with @method
    for name, obj in inspect.getmembers(handlers, inspect.iscoroutinefunction):
        # Skip private methods and helper functions
        if name.startswith('_') or name == 'register_methods':
            continue
        
        # Extract method information
        method_info = extract_method_info(obj)
        if method_info:
            # Check if method has a custom name (for rpc.discover)
            method_name = method_info["name"]
            if method_name == "rpc_discover":
                method_name = "rpc.discover"
            
            methods_list.append({
                "name": method_name,
                "description": method_info["description"],
                "params": method_info["params"],
                "result": method_info["result"],
                "errors": [
                    {
                        "$ref": "#/components/errors/ParseError"
                    },
                    {
                        "$ref": "#/components/errors/InvalidRequest"
                    },
                    {
                        "$ref": "#/components/errors/MethodNotFound"
                    },
                    {
                        "$ref": "#/components/errors/InvalidParams"
                    },
                    {
                        "$ref": "#/components/errors/InternalError"
                    },
                    {
                        "$ref": "#/components/errors/NotFoundError"
                    },
                    {
                        "$ref": "#/components/errors/ValidationError"
                    }
                ]
            })
    
    # Sort methods alphabetically for consistency, but keep rpc.discover first
    methods_list.sort(key=lambda x: (x["name"] != "rpc.discover", x["name"]))
    
    spec = {
        "openrpc": OPENRPC_VERSION,
        "info": {
            "title": SERVICE_NAME,
            "version": SERVICE_VERSION,
            "description": "Database-as-a-Service (DBaaS) with JSON-RPC 2.0 API. Provides multi-tenant data storage with flexible node and relationship management.",
            "contact": {
                "name": "flex-db",
                "url": "https://github.com/hemanthpathath/flex-db"
            }
        },
        "servers": [
            {
                "name": "Local Development",
                "url": "http://localhost:5000/jsonrpc",
                "description": "Local development server (use port 5001 when running via Docker)"
            },
            {
                "name": "Docker",
                "url": "http://localhost:5001/jsonrpc",
                "description": "Docker containerized service"
            }
        ],
        "methods": methods_list,
        "components": {
            "errors": {
                "ParseError": {
                    "code": -32700,
                    "message": "Parse error",
                    "data": {
                        "type": "string",
                        "description": "Error details"
                    }
                },
                "InvalidRequest": {
                    "code": -32600,
                    "message": "Invalid Request",
                    "data": {
                        "type": "string",
                        "description": "Error details"
                    }
                },
                "MethodNotFound": {
                    "code": -32601,
                    "message": "Method not found",
                    "data": {
                        "type": "string",
                        "description": "Error details"
                    }
                },
                "InvalidParams": {
                    "code": -32602,
                    "message": "Invalid params",
                    "data": {
                        "type": "string",
                        "description": "Error details"
                    }
                },
                "InternalError": {
                    "code": -32603,
                    "message": "Internal error",
                    "data": {
                        "type": "string",
                        "description": "Error details"
                    }
                },
                "NotFoundError": {
                    "code": -32001,
                    "message": "Resource not found",
                    "data": {
                        "type": "string",
                        "description": "Error details"
                    }
                },
                "ValidationError": {
                    "code": -32002,
                    "message": "Validation error",
                    "data": {
                        "type": "string",
                        "description": "Error details"
                    }
                }
            }
        }
    }
    
    return spec


def get_openrpc_spec_json() -> str:
    """Get OpenRPC spec as JSON string."""
    spec = generate_openrpc_spec()
    return json.dumps(spec, indent=2)
