# 字段命名规范（全链路）

本项目协议字段统一使用 **camelCase**，物品标识统一使用：

- `itemId`：物品唯一标识（如 `minecraft:iron_ingot`）
- `itemName`：展示名称（本地化后的人类可读名称）

## 1. 通用规则

- JSON / WebSocket / HTTP 的字段名统一使用 **camelCase**。
- Go 结构体字段使用 **PascalCase**，并通过 `json` tag 映射到 camelCase。
- 不再使用以下历史写法作为对外协议字段：
  - `item`
  - `ItemId`
  - `name`（用于表示物品 ID 的场景）

## 2. 推荐映射

- Go 字段：`ItemID` ↔ JSON 字段：`itemId`
- Go 字段：`ItemName` ↔ JSON 字段：`itemName`

## 3. 消息规范

### 3.1 产线流量上报（production_flow）

```json
{
  "type": "production_flow",
  "id": "factory_1",
  "name": "Main Factory",
  "delta": 64,
  "itemId": "minecraft:iron_ingot"
}
```

### 3.2 合成指令（craft）

```json
{
  "type": "craft",
  "itemId": "minecraft:glass",
  "count": 256
}
```

### 3.3 可合成列表回传（craftables）

```json
{
  "type": "craftables",
  "id": "ae_hub_1",
  "requestId": "1739440000000000000",
  "craftables": [
    { "itemId": "minecraft:glass", "itemName": "玻璃" }
  ]
}
```

## 4. 例外说明

- `name` 字段允许用于“设备名 / 工厂名”等实体名称语义，不用于物品 ID。
- 与第三方 API 交互时，如果第三方固定使用 `name`（如 `bridge.getItem({name=...})`），仅在适配层内部转换，进入本项目协议后必须转换为 `itemId`。
