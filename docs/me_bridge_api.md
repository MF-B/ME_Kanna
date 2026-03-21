# me_bridge API 文档

## 约定
- 表达式写法：`函数名(参数) -> 返回值`
- 返回 `nil` 时通常伴随 `err` 字符串（`result, err` 形式）
- 部分返回可能包含重复引用（例如 `tags`），在 JSON 里会被替换成 `"<ref>"`

## 基本函数
- `isOnline() -> boolean`
- `isConnected() -> boolean`
- `getItem(filter: table) -> table | nil, err: string`
- `getFluid(filter: table) -> table | nil, err: string`
- `getChemical(filter: table) -> table | nil, err: string`
- `getItems(filter: table) -> table | nil, err: string`
- `getFluids(filter: table) -> table | nil, err: string`
- `getChemicals(filter: table) -> table | nil, err: string`
- `getCraftableItems(filter: table) -> table | nil, err: string`
- `getCraftableFluids(filter: table) -> table | nil, err: string`
- `getCells() -> table | nil, err: string`
- `getConfiguration() -> table | nil, err: string`
- `getName() -> string`
- `getDrives() -> table | nil, err: string`

### 返回示例
`getName()`:
```json
"ME Bridge"
```

`getConfiguration()`:

`getItem()` 返回示例:
```json
{
   "components": {},
   "count": 34,
   "displayName": "[Iron Ingot]",
   "fingerprint": "-8242565751158650622",
   "isCraftable": false,
   "maxStackSize": 64,
   "name": "minecraft:iron_ingot",
   "tags": [
      "minecraft:item/morered:red_alloyable_ingots",
      "minecraft:item/minecraft:beacon_payment_items",
      "minecraft:item/minecolonies:blacksmith_ingredient",
      "minecraft:item/minecraft:trim_materials",
      "minecraft:item/minecolonies:reduceable_ingredient",
      "minecraft:item/irons_jewelry:loot_handler/low_gearscore",
      "minecraft:item/silentgear:greedy_magnet_attracted",
      "minecraft:item/c:ingots/iron",
      "minecraft:item/minecolonies:blacksmith_product",
      "minecraft:item/c:ingots",
      "minecraft:item/irons_spellbooks:arcane_ingot_base",
      "minecraft:item/minecolonies:reduceable_product_excluded",
      "minecraft:item/mekanism:muffling_center",
      "minecraft:item/minecolonies:sawmill_ingredient_excluded",
      "minecraft:item/theurgy:metals/mercury/low",
      "minecraft:item/ae2:metal_ingots"
   ]
}
```
```json
{}
```

`getCells()` (数组元素示例):
```json
{
   "bytes": 1024,
   "bytesPerType": 8,
   "fuzzyMode": "IGNORE_ALL",
   "item": {
      "name": "ae2:item_storage_cell_1k",
      "tags": "<ref>"
   },
   "totalTypes": 63,
   "type": "ae2:i",
   "usedBytes": 11
}
```

`getCraftableItems()` (数组元素示例):
```json
{
   "components": {},
   "count": 11,
   "displayName": "[Fluix Dust]",
   "fingerprint": "-4941552395787903306",
   "isCraftable": true,
   "maxStackSize": 64,
   "name": "ae2:fluix_dust",
   "tags": [
      "minecraft:item/c:dusts/fluix",
      "minecraft:item/supplementaries:hourglass_dusts",
      "minecraft:item/minecolonies:reduceable_ingredient",
      "minecraft:item/c:dusts"
   ]
}
```

`getDrives()` (数组元素示例):
```json
{
   "cells": [
      { "bytes": 1024, "bytesPerType": 8, "fuzzyMode": "IGNORE_ALL", "item": {"name": "ae2:item_storage_cell_1k", "tags": "<ref>"}, "totalTypes": 63, "type": "ae2:i", "usedBytes": 11 }
   ],
   "menuIcon": { "name": "ae2:drive", "tags": {} },
   "name": "ME Drive",
   "position": { "x": 1580, "y": 64, "z": -31 },
   "priority": 200,
   "totalBytes": 6144,
   "usedBytes": 692
}
```

## 输入/输出函数
- `importItem(filter: table) -> table | nil, err: string`
- `exportItem(filter: table) -> table | nil, err: string`
- `importFluid(filter: table) -> table | nil, err: string`
- `exportFluid(filter: table) -> table | nil, err: string`
- `importChemical(filter: table) -> table | nil, err: string`
- `exportChemical(filter: table) -> table | nil, err: string`

## 能量相关函数
- `getStoredEnergy() -> number`
- `getEnergyCapacity() -> number`
- `getEnergyUsage() -> number`
- `getAverageEnergyInput() -> number`

## 存储相关函数
- `getTotalExternalItemStorage() -> number`
- `getTotalExternalFluidStorage() -> number`
- `getTotalExternalChemicalStorage() -> number`
- `getTotalItemStorage() -> number`
- `getTotalFluidStorage() -> number`
- `getTotalChemicalStorage() -> number`
- `getUsedExternalItemStorage() -> number`
- `getUsedExternalFluidStorage() -> number`
- `getUsedExternalChemicalStorage() -> number`
- `getUsedItemStorage() -> number`
- `getUsedFluidStorage() -> number`
- `getUsedChemicalStorage() -> number`
- `getAvailableExternalItemStorage() -> number`
- `getAvailableExternalFluidStorage() -> number`
- `getAvailableExternalChemicalStorage() -> number`
- `getAvailableItemStorage() -> number`
- `getAvailableFluidStorage() -> number`
- `getAvailableChemicalStorage() -> number`

## 合成相关函数
事件字段：
- `error: boolean` 计算或合成是否失败
- `id: int` 合成任务 id
- `debug_message: string` 任务状态描述

```lua
local event, error, id, message = os.pullEvent("ae_crafting")
print("A crafting update occurred for Job #" .. id)
if error then
      print("There was an error while calculating or crafting the resource with the message " .. message)
else
      print("The new state of the task is " .. message)
end
```

函数列表：
- `craftItem(filter: table) -> table | nil, err: string`
- `craftFluid(filter: table) -> table | nil, err: string`
- `getCraftingTask(id: int) -> table | nil, err: string`
- `getCraftingTasks() -> table`
- `getCraftingCPUs() -> table, err: string`
- `cancelCraftingTasks(filter: table) -> int`
- `getPatterns(pattern_filter: table) -> table | nil, err: string`
- `isCraftable(filter: table) -> boolean`
- `isCrafting(filter: table) -> boolean`

### 返回示例
`getCraftingTasks()`（有任务时，数组元素示例，注意此函数不显示玩家下单的任务）:
```json
{
   "bridge_id": 373,
   "completion": 0.25,
   "cpu": {
      "coProcessors": 1,
      "isBusy": true,
      "name": "Unnamed",
      "selectionMode": "ANY",
      "storage": 81920
   },
   "crafted": 4,
   "id": "35045d52-5023-4dc0-9f6e-1377c880779d",
   "quantity": 16,
   "resource": {
      "components": {},
      "count": 16,
      "displayName": "[Certus Quartz Dust]",
      "fingerprint": "7244626626843091630",
      "isCraftable": false,
      "maxStackSize": 64,
      "name": "ae2:certus_quartz_dust",
      "tags": [
         "minecraft:item/c:dusts",
         "minecraft:item/ae2:all_quartz_dust",
         "minecraft:item/supplementaries:hourglass_dusts",
         "minecraft:item/c:dusts/certus_quartz",
         "minecraft:item/minecolonies:reduceable_ingredient"
      ]
   }
}
```

`getCraftingCPUs()` (数组元素示例):
```json
{
   "coProcessors": 1,
   "craftingJob": {
      "bridge_id": -1,
      "quantity": 16,
      "resource": {
         "components": {},
         "count": 16,
         "displayName": "[Certus Quartz Dust]",
         "fingerprint": "7244626626843091630",
         "isCraftable": false,
         "maxStackSize": 64,
         "name": "ae2:certus_quartz_dust",
         "tags": [
            "minecraft:item/c:dusts",
            "minecraft:item/ae2:all_quartz_dust",
            "minecraft:item/supplementaries:hourglass_dusts",
            "minecraft:item/c:dusts/certus_quartz",
            "minecraft:item/minecolonies:reduceable_ingredient"
         ]
      }
   },
   "isBusy": true,
   "name": "Unnamed",
   "selectionMode": "ANY",
   "storage": 65536
}
```

`getPatterns()` (数组元素示例):
```json
{
   "inputs": [
      {
         "multiplier": 5,
         "primaryInput": {
            "components": {},
            "count": 1,
            "displayName": "[Certus Quartz Dust]",
            "fingerprint": "7244626626843091630",
            "isCraftable": false,
            "maxStackSize": 64,
            "name": "ae2:certus_quartz_dust",
            "tags": [
               "minecraft:item/c:dusts",
               "minecraft:item/ae2:all_quartz_dust",
               "minecraft:item/supplementaries:hourglass_dusts",
               "minecraft:item/c:dusts/certus_quartz",
               "minecraft:item/minecolonies:reduceable_ingredient"
            ]
         }
      }
   ],
   "outputs": [
      {
         "components": {},
         "count": 4,
         "displayName": "[Quartz Glass]",
         "fingerprint": "-3616077736511432026",
         "isCraftable": false,
         "maxStackSize": 64,
         "name": "ae2:quartz_glass",
         "tags": "<ref>"
      }
   ],
   "patternType": "crafting",
   "primaryOutput": {
      "components": {},
      "count": 4,
      "displayName": "[Quartz Glass]",
      "fingerprint": "-3616077736511432026",
      "isCraftable": false,
      "maxStackSize": 64,
      "name": "ae2:quartz_glass",
      "tags": [
         "minecraft:item/minecolonies:glassblower_smelting_product",
         "minecraft:item/minecolonies:glassblower_ingredient",
         "minecraft:item/minecolonies:reduceable_ingredient",
         "minecraft:item/c:glass_blocks"
      ]
   }
}
```

## getPatterns 过滤器
```lua
-- 输出包含木棍
outputFilter = { output = { name = "minecraft:stick" } }

-- 输入包含木棍
inputFilter = { input = { name = "minecraft:stick" } }

-- 输出木棍 + 输入木板
inputFilter = {
      input = { name = "minecraft:oak_planks" },
      output = { name = "minecraft:stick" }
}

-- 不传参数，返回全部
patterns = bridge.getPatterns()
```
