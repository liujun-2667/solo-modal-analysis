<template>
    <div class="structure-form">
        <el-tabs v-model="activeTab">
            <el-tab-pane label="节点" name="nodes">
                <div class="tab-content">
                    <el-button @click="addNode" type="primary" size="small">添加节点</el-button>
                    <el-table :data="nodes" border>
                        <el-table-column label="编号" prop="id" />
                        <el-table-column label="X" prop="x">
                            <template #default="scope">
                                <el-input-number v-model="scope.row.x" :step="0.1" />
                            </template>
                        </el-table-column>
                        <el-table-column label="Y" prop="y">
                            <template #default="scope">
                                <el-input-number v-model="scope.row.y" :step="0.1" />
                            </template>
                        </el-table-column>
                        <el-table-column label="Z" prop="z">
                            <template #default="scope">
                                <el-input-number v-model="scope.row.z" :step="0.1" />
                            </template>
                        </el-table-column>
                        <el-table-column>
                            <template #default="scope">
                                <el-button @click="removeNode(scope.$index)" type="danger" size="small">删除</el-button>
                            </template>
                        </el-table-column>
                    </el-table>
                </div>
            </el-tab-pane>
            <el-tab-pane label="单元" name="elements">
                <div class="tab-content">
                    <el-button @click="addElement" type="primary" size="small">添加单元</el-button>
                    <el-table :data="elements" border>
                        <el-table-column label="编号" prop="id" />
                        <el-table-column label="节点1">
                            <template #default="scope">
                                <el-select v-model="scope.row.node1">
                                    <el-option v-for="n in nodes" :key="n.id" :label="n.id" :value="n.id" />
                                </el-select>
                            </template>
                        </el-table-column>
                        <el-table-column label="节点2">
                            <template #default="scope">
                                <el-select v-model="scope.row.node2">
                                    <el-option v-for="n in nodes" :key="n.id" :label="n.id" :value="n.id" />
                                </el-select>
                            </template>
                        </el-table-column>
                        <el-table-column label="截面">
                            <template #default="scope">
                                <el-select v-model="scope.row.sectionId">
                                    <el-option v-for="s in sections" :key="s.id" :label="s.id" :value="s.id" />
                                </el-select>
                            </template>
                        </el-table-column>
                        <el-table-column>
                            <template #default="scope">
                                <el-button @click="removeElement(scope.$index)" type="danger" size="small">删除</el-button>
                            </template>
                        </el-table-column>
                    </el-table>
                </div>
            </el-tab-pane>
            <el-tab-pane label="截面属性" name="sections">
                <div class="tab-content">
                    <el-button @click="addSection" type="primary" size="small">添加截面</el-button>
                    <el-table :data="sections" border>
                        <el-table-column label="编号" prop="id" />
                        <el-table-column label="面积A">
                            <template #default="scope">
                                <el-input-number v-model="scope.row.a" :step="0.001" />
                            </template>
                        </el-table-column>
                        <el-table-column label="惯性矩Ix">
                            <template #default="scope">
                                <el-input-number v-model="scope.row.ix" :step="1e-8" />
                            </template>
                        </el-table-column>
                        <el-table-column label="惯性矩Iy">
                            <template #default="scope">
                                <el-input-number v-model="scope.row.iy" :step="1e-8" />
                            </template>
                        </el-table-column>
                        <el-table-column label="惯性矩Iz">
                            <template #default="scope">
                                <el-input-number v-model="scope.row.iz" :step="1e-8" />
                            </template>
                        </el-table-column>
                        <el-table-column label="弹性模量E">
                            <template #default="scope">
                                <el-input-number v-model="scope.row.e" :step="1e10" />
                            </template>
                        </el-table-column>
                        <el-table-column label="密度rho">
                            <template #default="scope">
                                <el-input-number v-model="scope.row.rho" :step="100" />
                            </template>
                        </el-table-column>
                        <el-table-column label="泊松比nu">
                            <template #default="scope">
                                <el-input-number v-model="scope.row.nu" :step="0.01" :min="0" :max="0.5" />
                            </template>
                        </el-table-column>
                        <el-table-column>
                            <template #default="scope">
                                <el-button @click="removeSection(scope.$index)" type="danger" size="small">删除</el-button>
                            </template>
                        </el-table-column>
                    </el-table>
                </div>
            </el-tab-pane>
            <el-tab-pane label="约束条件" name="constraints">
                <div class="tab-content">
                    <el-button @click="addConstraint" type="primary" size="small">添加约束</el-button>
                    <el-table :data="constraints" border>
                        <el-table-column label="节点">
                            <template #default="scope">
                                <el-select v-model="scope.row.nodeId">
                                    <el-option v-for="n in nodes" :key="n.id" :label="n.id" :value="n.id" />
                                </el-select>
                            </template>
                        </el-table-column>
                        <el-table-column label="DX">
                            <template #default="scope">
                                <el-switch v-model="scope.row.dx" />
                            </template>
                        </el-table-column>
                        <el-table-column label="DY">
                            <template #default="scope">
                                <el-switch v-model="scope.row.dy" />
                            </template>
                        </el-table-column>
                        <el-table-column label="DZ">
                            <template #default="scope">
                                <el-switch v-model="scope.row.dz" />
                            </template>
                        </el-table-column>
                        <el-table-column label="RX">
                            <template #default="scope">
                                <el-switch v-model="scope.row.rx" />
                            </template>
                        </el-table-column>
                        <el-table-column label="RY">
                            <template #default="scope">
                                <el-switch v-model="scope.row.ry" />
                            </template>
                        </el-table-column>
                        <el-table-column label="RZ">
                            <template #default="scope">
                                <el-switch v-model="scope.row.rz" />
                            </template>
                        </el-table-column>
                        <el-table-column>
                            <template #default="scope">
                                <el-button @click="removeConstraint(scope.$index)" type="danger" size="small">删除</el-button>
                            </template>
                        </el-table-column>
                    </el-table>
                </div>
            </el-tab-pane>
        </el-tabs>
        <div class="footer">
            <el-button @click="loadExample" type="info">加载预设算例</el-button>
            <el-button @click="exportModel">导出JSON</el-button>
            <el-button @click="importModel">导入JSON</el-button>
            <el-button @click="clearModel">清空模型</el-button>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Node, Element, Section, Constraint } from '../types'

const emit = defineEmits<{
    (e: 'update', data: { nodes: Node[], elements: Element[], sections: Section[], constraints: Constraint[] }): void
}>()

const activeTab = ref('nodes')

const nodes = ref<Node[]>([])
const elements = ref<Element[]>([])
const sections = ref<Section[]>([])
const constraints = ref<Constraint[]>([])

const addNode = () => {
    const id = nodes.value.length > 0 ? Math.max(...nodes.value.map(n => n.id)) + 1 : 1
    nodes.value.push({ id, x: 0, y: 0, z: 0 })
    emitUpdate()
}

const removeNode = (index: number) => {
    nodes.value.splice(index, 1)
    emitUpdate()
}

const addElement = () => {
    const id = elements.value.length > 0 ? Math.max(...elements.value.map(e => e.id)) + 1 : 1
    const node1 = nodes.value.length > 0 ? nodes.value[0].id : 1
    const node2 = nodes.value.length > 1 ? nodes.value[1].id : 2
    const sectionId = sections.value.length > 0 ? sections.value[0].id : 1
    elements.value.push({ id, node1, node2, sectionId })
    emitUpdate()
}

const removeElement = (index: number) => {
    elements.value.splice(index, 1)
    emitUpdate()
}

const addSection = () => {
    const id = sections.value.length > 0 ? Math.max(...sections.value.map(s => s.id)) + 1 : 1
    sections.value.push({
        id,
        a: 0.01,
        ix: 8.333e-6,
        iy: 1e-6,
        iz: 1e-6,
        e: 2.1e11,
        rho: 7850,
        nu: 0.3
    })
    emitUpdate()
}

const removeSection = (index: number) => {
    sections.value.splice(index, 1)
    emitUpdate()
}

const addConstraint = () => {
    const nodeId = nodes.value.length > 0 ? nodes.value[0].id : 1
    constraints.value.push({
        nodeId,
        dx: false,
        dy: false,
        dz: false,
        rx: false,
        ry: false,
        rz: false
    })
    emitUpdate()
}

const removeConstraint = (index: number) => {
    constraints.value.splice(index, 1)
    emitUpdate()
}

const loadExample = () => {
    emit('update', { nodes: [], elements: [], sections: [], constraints: [] })
}

const exportModel = () => {
    const model = {
        nodes: nodes.value,
        elements: elements.value,
        sections: sections.value,
        constraints: constraints.value
    }
    const json = JSON.stringify(model, null, 2)
    const blob = new Blob([json], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'structure_model.json'
    a.click()
    URL.revokeObjectURL(url)
}

const importModel = () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.json'
    input.onchange = (e) => {
        const file = (e.target as HTMLInputElement).files?.[0]
        if (file) {
            const reader = new FileReader()
            reader.onload = (event) => {
                try {
                    const model = JSON.parse(event.target?.result as string)
                    nodes.value = model.nodes || []
                    elements.value = model.elements || []
                    sections.value = model.sections || []
                    constraints.value = model.constraints || []
                    emitUpdate()
                } catch (error) {
                    console.error('Failed to parse JSON:', error)
                }
            }
            reader.readAsText(file)
        }
    }
    input.click()
}

const clearModel = () => {
    nodes.value = []
    elements.value = []
    sections.value = []
    constraints.value = []
    emitUpdate()
}

const emitUpdate = () => {
    emit('update', {
        nodes: nodes.value,
        elements: elements.value,
        sections: sections.value,
        constraints: constraints.value
    })
}

defineProps<{
    initialNodes?: Node[]
    initialElements?: Element[]
    initialSections?: Section[]
    initialConstraints?: Constraint[]
}>()
</script>

<style scoped>
.structure-form {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.tab-content {
    flex: 1;
    overflow: auto;
}

.footer {
    padding: 10px;
    border-top: 1px solid #e8e8e8;
    display: flex;
    gap: 10px;
}
</style>
