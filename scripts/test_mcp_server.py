#!/usr/bin/env python3
"""
Simple MCP Server for testing AgentFlow integration.
This server provides basic file operations and demonstrates MCP protocol.
"""

import json
import sys
from typing import Any, Dict, List, Optional
import asyncio
import os
from datetime import datetime

print("PYTHON CWD:", os.getcwd(), flush=True)

class SimpleMCPServer:
    def __init__(self):
        self.protocol_version = "2024-11-05"
        self.server_info = {
            "name": "agentflow-test-server",
            "version": "1.0.0"
        }
        self.capabilities = {
            "tools": {}
        }

    async def handle_initialize(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """Handle initialization request."""
        return {
            "protocolVersion": self.protocol_version,
            "capabilities": self.capabilities,
            "serverInfo": self.server_info
        }

    async def handle_list_tools(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """List available tools."""
        tools = [
            {
                "name": "read_file",
                "description": "Read the contents of a file",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": "Path to the file to read"
                        }
                    },
                    "required": ["path"]
                }
            },
            {
                "name": "write_file",
                "description": "Write content to a file",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": "Path to the file to write"
                        },
                        "content": {
                            "type": "string",
                            "description": "Content to write to the file"
                        }
                    },
                    "required": ["path", "content"]
                }
            },
            {
                "name": "list_directory",
                "description": "List contents of a directory",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": "Path to the directory to list"
                        }
                    },
                    "required": ["path"]
                }
            },
            {
                "name": "get_timestamp",
                "description": "Get current timestamp",
                "inputSchema": {
                    "type": "object",
                    "properties": {},
                    "required": []
                }
            }
        ]
        
        return {"tools": tools}

    async def handle_call_tool(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """Handle tool call request."""
        tool_name = params.get("name")
        arguments = params.get("arguments", {})

        try:
            if tool_name == "read_file":
                return await self._read_file(arguments)
            elif tool_name == "write_file":
                return await self._write_file(arguments)
            elif tool_name == "list_directory":
                return await self._list_directory(arguments)
            elif tool_name == "get_timestamp":
                return await self._get_timestamp(arguments)
            else:
                return {
                    "content": [
                        {
                            "type": "text",
                            "text": f"Unknown tool: {tool_name}"
                        }
                    ],
                    "isError": True
                }
        except Exception as e:
            return {
                "content": [
                    {
                        "type": "text",
                        "text": f"Error executing tool {tool_name}: {str(e)}"
                    }
                ],
                "isError": True
            }

    async def _read_file(self, args: Dict[str, Any]) -> Dict[str, Any]:
        """Read file tool implementation."""
        path = args.get("path")
        if not path:
            raise ValueError("path is required")

        try:
            with open(path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            return {
                "content": [
                    {
                        "type": "text",
                        "text": f"File content from {path}:\n{content}"
                    }
                ]
            }
        except FileNotFoundError:
            return {
                "content": [
                    {
                        "type": "text",
                        "text": f"File not found: {path}"
                    }
                ],
                "isError": True
            }

    async def _write_file(self, args: Dict[str, Any]) -> Dict[str, Any]:
        """Write file tool implementation."""
        path = args.get("path")
        content = args.get("content")
        
        if not path:
            raise ValueError("path is required")
        if content is None:
            raise ValueError("content is required")

        try:
            os.makedirs(os.path.dirname(path), exist_ok=True)
            with open(path, 'w', encoding='utf-8') as f:
                f.write(content)
            
            return {
                "content": [
                    {
                        "type": "text",
                        "text": f"Successfully wrote {len(content)} characters to {path}"
                    }
                ]
            }
        except Exception as e:
            return {
                "content": [
                    {
                        "type": "text",
                        "text": f"Failed to write file: {str(e)}"
                    }
                ],
                "isError": True
            }

    async def _list_directory(self, args: Dict[str, Any]) -> Dict[str, Any]:
        """List directory tool implementation."""
        path = args.get("path", ".")
        
        try:
            if not os.path.exists(path):
                return {
                    "content": [
                        {
                            "type": "text",
                            "text": f"Directory not found: {path}"
                        }
                    ],
                    "isError": True
                }

            if not os.path.isdir(path):
                return {
                    "content": [
                        {
                            "type": "text",
                            "text": f"Path is not a directory: {path}"
                        }
                    ],
                    "isError": True
                }

            items = []
            for item in sorted(os.listdir(path)):
                item_path = os.path.join(path, item)
                item_type = "directory" if os.path.isdir(item_path) else "file"
                size = os.path.getsize(item_path) if os.path.isfile(item_path) else 0
                items.append(f"{item} ({item_type}, {size} bytes)")

            content = f"Contents of {path}:\n" + "\n".join(items)
            
            return {
                "content": [
                    {
                        "type": "text",
                        "text": content
                    }
                ]
            }
        except Exception as e:
            return {
                "content": [
                    {
                        "type": "text",
                        "text": f"Failed to list directory: {str(e)}"
                    }
                ],
                "isError": True
            }

    async def _get_timestamp(self, args: Dict[str, Any]) -> Dict[str, Any]:
        """Get timestamp tool implementation."""
        timestamp = datetime.now().isoformat()
        
        return {
            "content": [
                {
                    "type": "text",
                    "text": f"Current timestamp: {timestamp}"
                }
            ]
        }

    async def handle_message(self, message: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Handle incoming MCP message."""
        method = message.get("method")
        params = message.get("params", {})
        request_id = message.get("id")

        response = None

        if method == "initialize":
            result = await self.handle_initialize(params)
            response = {
                "jsonrpc": "2.0",
                "id": request_id,
                "result": result
            }
        elif method == "tools/list":
            result = await self.handle_list_tools(params)
            response = {
                "jsonrpc": "2.0",
                "id": request_id,
                "result": result
            }
        elif method == "tools/call":
            result = await self.handle_call_tool(params)
            response = {
                "jsonrpc": "2.0",
                "id": request_id,
                "result": result
            }
        else:
            response = {
                "jsonrpc": "2.0",
                "id": request_id,
                "error": {
                    "code": -32601,
                    "message": f"Method not found: {method}"
                }
            }

        return response

    async def run(self):
        """Run the MCP server."""
        try:
            while True:
                line = await asyncio.get_event_loop().run_in_executor(None, sys.stdin.readline)
                if not line:
                    break

                line = line.strip()
                if not line:
                    continue

                try:
                    message = json.loads(line)
                    response = await self.handle_message(message)
                    if response:
                        print(json.dumps(response), flush=True)
                except json.JSONDecodeError as e:
                    error_response = {
                        "jsonrpc": "2.0",
                        "id": None,
                        "error": {
                            "code": -32700,
                            "message": f"Parse error: {str(e)}"
                        }
                    }
                    print(json.dumps(error_response), flush=True)
                except Exception as e:
                    error_response = {
                        "jsonrpc": "2.0",
                        "id": None,
                        "error": {
                            "code": -32603,
                            "message": f"Internal error: {str(e)}"
                        }
                    }
                    print(json.dumps(error_response), flush=True)

        except KeyboardInterrupt:
            pass
        except Exception as e:
            sys.stderr.write(f"Server error: {str(e)}\n")
            sys.stderr.flush()

if __name__ == "__main__":
    server = SimpleMCPServer()
    asyncio.run(server.run())
