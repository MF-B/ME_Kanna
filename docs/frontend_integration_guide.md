# 前后端对接需求清单 (Frontend-Backend Integration Checklist)

如果你准备为本项目开发一个新的前端（例如移动端、小程序或其他 Web 框架版本），你需要向后端开发人员确认以下信息。

## 1. 基础连接信息 (Base Connection)

*   **API 基础地址 (Base URL)**: 
    *   当前默认: `http://<host>:8080` (HTTP) 和 `ws://<host>:8080/ws/web` (WebSocket)。
    *   需要确认: 生产环境或开发环境的实际 IP/域名及端口。
*   **协议类型**: 
    *   当前使用 HTTP/1.1 和标准 WebSocket。
    *   需要确认: 是否支持 HTTPS/WSS，是否需要处理跨域 (CORS) 限制。

## 2. 实时数据推送 (WebSocket)

本项目核心数据通过 WebSocket 实时同步，需确认：

*   **连接路径**: 确认具体的 WS 路由（如 `/ws/web`）。
*   **心跳机制**: 后端是否要求发送 PING/PONG 包以维持连接。
*   **消息包格式**: 
    *   当前格式: JSON 字符串。
    *   关键字段: `type` (消息类型，如 `update`)，`data` (工厂列表)，`system` (系统全局状态)。
*   **重连机制**: 服务端是否有连接数限制，断线后是否需要重新订阅特定频道。

## 3. 业务接口 (HTTP API)

主要用于非实时、指令性的操作（如合成任务管理）：

*   **接口规范**: 确认是否遵循 RESTful 规范。
*   **常用接口清单**:
    *   `GET /autocraft/craftables`: 获取所有可合成物品列表。
    *   `GET /autocraft/tasks`: 获取当前正在进行的自动合成任务。
    *   `POST /autocraft/tasks`: 创建或更新任务（需确认 Payload 结构）。
    *   `DELETE /autocraft/tasks/:itemId`: 删除指定任务。
    *   `PATCH /autocraft/tasks/:itemId`: 局部更新任务状态（如开关 `isActive`）。
*   **数据结构 (Schema)**:
    *   请求体字段名（例如是 `itemId` 还是 `id`）。
    *   响应体结构（例如是否统一包装在 `{ code, data, msg }` 中）。

## 4. 鉴权与安全 (Auth)

*   **身份验证**: 
    *   当前前端代码中未见显式的 Token 或 Cookie 处理（可能是局域网直连或由网关处理）。
    *   需要确认: 是否需要 `Authorization` 请求头？是否有登录接口？
*   **白名单/权限**: 后端是否对来源 IP 有白名单限制。

## 5. 数据字典与枚举 (Data Dictionary)

*   **物品 ID 映射**: `itemId` (如 `minecraft:iron_ingot`) 对应的显示名称、图标获取路径。
*   **状态枚举**: `SystemStatus` 的各个数值代表的含义（如能量单位、存储百分比的计算方式）。

## 6. 错误处理 (Error Handling)

*   **错误码定义**: 确认非 200 状态码时的错误响应格式。
*   **超时设置**: 后端建议的请求超时时间（当前前端默认为 8s）。

---
*本文档基于 `mineCCT-web` 源代码总结，生成于 2026-02-15。*
