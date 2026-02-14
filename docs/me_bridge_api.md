# me_bridge包含的方法
## 基本函数
isOnline
isConnected
getItem(filter: table) -> table | nil, string[Returns the first item that matches the filter. Or nil with a debug message if none could be found]
getFluid
getChemical
getItems(filter: table) -> table | nil, string[Returns every item that matches the filter or an empty table if none could be found. An empty filter can be provided to get every resource. Returns nil if there was an issue parsing your table]
getFluids
getChemicals
getCraftableItems(filter: table) -> table | nil, string[Returns every craftable item that matches the filter even if there is currently no item stored or an empty table if none could be found. An empty filter can be provided to get every resource. Returns nil if there was an issue parsing your table.]
getCraftableFluids
getCells() -> table | nil, string[Returns every storage cell in the drives of the grid. Supports standard RS/ME cells and some third party cells. Please open a feature request if some custom addon cells do not work]
getConfiguration
getName
getCells() -> table | nil, string[Returns every drive connected to the system with the cells in it.]

## 输入/输出函数
importItem
exportItem
importFluid
exportFluid
importChemical
exportChemical

# 能量相关函数
getStoredEnergy
getEnergyCapacity
getEnergyUsage
getAverageEnergyInput

# 存储相关函数
getTotalExternalItemStorage
getTotalExternalFluidStorage
getTotalExternalChemicalStorage
getTotalItemStorage
getTotalFluidStorage
getTotalChemicalStorage
getUsedExternalItemStorage
getUsedExternalFluidStorage
getUsedExternalChemicalStorage
getUsedItemStorage
getUsedFluidStorage
getUsedChemicalStorage
getAvailableExternalItemStorage
getAvailableExternalFluidStorage
getAvailableExternalChemicalStorage
getAvailableItemStorage
getAvailableFluidStorage 
getAvailableChemicalStorage

# 合成相关函数
The new crafting event will be fired when the state of a task will change.

The name of the event is prefixed depending if you use the RS or ME Bridge. Use rs_crafting for the RS Bridge and me_crafting for the ME Bridge. Values:

1. error: boolean If an error occurred and the calculation or crafting was not successful
2. id: int The id of the craft job. Can be used to get the craft object
3. debug_message: string A debug message describing the current task of the task
```lua
local event, error, id, message = os.pullEvent("rs_crafting")
print("A crafting update occurred for Job #" .. id)
if error then
    print("There was an error while calculating or crafting the resource with the message " .. message)
else 
    print("The new state of the task is " .. message)
end
```

craftItem(filter: table) -> table | nil, string[Schedules a craft job for items. Will fire the crafting event when changes occur. Or nil if there was an issue with parsing the filter]
craftFluid(filter: table) -> table | nil, string[Schedules a craft job for fluids. Will fire the crafting event when changes occur. Or nil if there was an issue with parsing the filter]
getCraftingTask(id: int) -> table | nil, string[Returns the Crafting Job Object with the id. Nil if no object could be found]
getCraftingTasks() -> table[Returns every crafting task that is currently running.]
getCraftingCPUs() -> table, err: string
cancelCraftingTasks(filter: table) -> int[Cancels every crafting task where the output matches the filter]
getPatterns(pattern_filter: table) -> table | nil, string
[
Returns every pattern available to the grid or nil if there was an issue with parsing one of the filters. The filter of this function is a bit different
You can specify filters for the output and input of patterns. If you don't provide any filter or just no argument at all, it will return every pattern.
To specify an input or output filter, you need to use an input/output key in the filters table.
```lua
-- Filter for any patterns with sticks as an output
outputFilter = {
     output = { 
        name="minecraft:stick" 
     } 
}

-- Filter for any patterns with sticks as an input
inputFilter = {
     input = { 
        name="minecraft:stick" 
     } 
}

-- Filter for any patterns with sticks as an output and planks as an input
inputFilter = {
     input = {
        name = "minecraft:oak_planks"
     },
     output = { 
        name="minecraft:stick" 
     } 
}

-- Or use the function without argument to get every pattern
patterns = bridge.getPatterns()
```
]
isCraftable(filter: table) -> boolean
isCrafting(filter: table) -> boolean
