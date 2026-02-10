import { ref } from 'vue'

// 使用全局状态，这样 App.vue 和 ItemIcon.vue 都能读到同一个设置
// 优先从浏览器缓存(localStorage)里读取，默认用 'vanilla'
const currentPack = ref(localStorage.getItem('texture-pack') || 'vanilla')

export function useTexturePack() {
  
  // 切换材质包的函数
  const setPack = (packName) => {
    currentPack.value = packName
    // 保存到浏览器缓存，刷新页面不丢失
    localStorage.setItem('texture-pack', packName)
  }

  // 定义可用的材质包列表 (对应 el-select 的选项)
  // value 必须对应后端 mods/ 下面的 jar 包里的资源路径，或者你 public/packs 下的文件夹名
  // 这里我们其实主要是在切换 "vanilla" (原版) 和其他
  const availablePacks = [
    { label: '原版 (Vanilla)', value: 'vanilla' },
    // 如果你有其他的，可以加在这里，比如:
    // { label: '高清 (Faithful)', value: 'faithful' },
  ]

  return {
    currentPack,
    setPack,
    availablePacks
  }
}
