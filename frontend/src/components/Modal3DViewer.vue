<template>
    <div class="viewer-container">
        <div ref="container" class="canvas-container"></div>
        <div class="controls">
            <div class="control-group">
                <label>振型缩放因子</label>
                <el-slider v-model="scaleFactor" :min="0.1" :max="10" :step="0.1" />
                <span>{{ scaleFactor.toFixed(1) }}x</span>
            </div>
            <div class="control-group">
                <label>动画速度</label>
                <el-slider v-model="animationSpeed" :min="0.1" :max="5" :step="0.1" />
                <span>{{ animationSpeed.toFixed(1) }}x</span>
            </div>
            <div class="control-group">
                <el-button @click="toggleAnimation" :icon="isAnimating ? Pause : Play">
                    {{ isAnimating ? '暂停' : '播放' }}
                </el-button>
            </div>
        </div>
        <div class="color-bar">
            <span>位移</span>
            <div class="color-gradient"></div>
            <span>大</span>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import * as THREE from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'
import { Play, Pause } from '@element-plus/icons-vue'
import type { Node, Element, ModalResult, ModeSelection } from '../types'

const props = defineProps<{
    nodes: Node[]
    elements: Element[]
    modalResults: ModalResult[]
    selectedMode: number
    modeSelections: ModeSelection[]
}>()

const container = ref<HTMLDivElement | null>(null)
const scaleFactor = ref(1)
const animationSpeed = ref(1)
const isAnimating = ref(true)

let scene: THREE.Scene
let camera: THREE.PerspectiveCamera
let renderer: THREE.WebGLRenderer
let controls: OrbitControls
let originalLine: THREE.Line
let deformedLine: THREE.Line
let animationId: number
let time = 0

const currentModeShape = computed(() => {
    if (props.selectedMode >= 0 && props.selectedMode < props.modalResults.length) {
        return props.modalResults[props.selectedMode].modeShape
    }
    return []
})

const combinedModeShape = computed(() => {
    if (props.modeSelections.filter(s => s.enabled).length === 0) {
        return currentModeShape.value
    }
    const shape = new Array(props.nodes.length * 6).fill(0)
    props.modeSelections.forEach(sel => {
        if (sel.enabled && sel.modeIndex < props.modalResults.length) {
            const modeShape = props.modalResults[sel.modeIndex].modeShape
            modeShape.forEach((val, i) => {
                shape[i] += val * sel.scale
            })
        }
    })
    return shape
})

const initScene = () => {
    if (!container.value) return

    scene = new THREE.Scene()
    scene.background = new THREE.Color(0xf5f5f5)

    camera = new THREE.PerspectiveCamera(60, container.value.clientWidth / container.value.clientHeight, 0.1, 1000)
    camera.position.set(15, 10, 15)

    renderer = new THREE.WebGLRenderer({ antialias: true })
    renderer.setSize(container.value.clientWidth, container.value.clientHeight)
    renderer.setPixelRatio(window.devicePixelRatio)
    container.value.appendChild(renderer.domElement)

    controls = new OrbitControls(camera, renderer.domElement)
    controls.enableDamping = true
    controls.dampingFactor = 0.05

    const ambientLight = new THREE.AmbientLight(0xffffff, 0.6)
    scene.add(ambientLight)

    const directionalLight = new THREE.DirectionalLight(0xffffff, 0.8)
    directionalLight.position.set(10, 20, 10)
    scene.add(directionalLight)

    const gridHelper = new THREE.GridHelper(30, 30, 0xcccccc, 0xcccccc)
    scene.add(gridHelper)

    createStructure()

    animate()
}

const createStructure = () => {
    if (!props.nodes.length || !props.elements.length) return

    if (originalLine) scene.remove(originalLine)
    if (deformedLine) scene.remove(deformedLine)

    const originalGeometry = new THREE.BufferGeometry()
    const deformedGeometry = new THREE.BufferGeometry()

    const originalPoints: THREE.Vector3[] = []
    const deformedPoints: THREE.Vector3[] = []

    props.elements.forEach(elem => {
        const node1 = props.nodes.find(n => n.id === elem.node1)
        const node2 = props.nodes.find(n => n.id === elem.node2)
        if (node1 && node2) {
            originalPoints.push(new THREE.Vector3(node1.x, node1.y, node1.z))
            originalPoints.push(new THREE.Vector3(node2.x, node2.y, node2.z))
            deformedPoints.push(new THREE.Vector3(node1.x, node1.y, node1.z))
            deformedPoints.push(new THREE.Vector3(node2.x, node2.y, node2.z))
        }
    })

    originalGeometry.setFromPoints(originalPoints)
    deformedGeometry.setFromPoints(deformedPoints)

    const originalMaterial = new THREE.LineBasicMaterial({ color: 0xaaaaaa, linewidth: 1 })
    const deformedMaterial = new THREE.LineBasicMaterial({ color: 0xff0000, linewidth: 3 })

    originalLine = new THREE.LineSegments(originalGeometry, originalMaterial)
    deformedLine = new THREE.LineSegments(deformedGeometry, deformedMaterial)

    scene.add(originalLine)
    scene.add(deformedLine)
}

const updateDeformedShape = (factor: number) => {
    if (!deformedLine || !props.nodes.length) return

    const positions = deformedLine.geometry.attributes.position.array as Float32Array
    const modeShape = combinedModeShape.value

    let pointIndex = 0
    props.elements.forEach(elem => {
        const node1 = props.nodes.find(n => n.id === elem.node1)
        const node2 = props.nodes.find(n => n.id === elem.node2)
        if (node1 && node2) {
            const nodeIdx1 = props.nodes.findIndex(n => n.id === elem.node1)
            const nodeIdx2 = props.nodes.findIndex(n => n.id === elem.node2)

            positions[pointIndex * 3] = node1.x + (modeShape[nodeIdx1 * 6] || 0) * scaleFactor.value * factor
            positions[pointIndex * 3 + 1] = node1.y + (modeShape[nodeIdx1 * 6 + 1] || 0) * scaleFactor.value * factor
            positions[pointIndex * 3 + 2] = node1.z + (modeShape[nodeIdx1 * 6 + 2] || 0) * scaleFactor.value * factor
            pointIndex++

            positions[pointIndex * 3] = node2.x + (modeShape[nodeIdx2 * 6] || 0) * scaleFactor.value * factor
            positions[pointIndex * 3 + 1] = node2.y + (modeShape[nodeIdx2 * 6 + 1] || 0) * scaleFactor.value * factor
            positions[pointIndex * 3 + 2] = node2.z + (modeShape[nodeIdx2 * 6 + 2] || 0) * scaleFactor.value * factor
            pointIndex++
        }
    })

    deformedLine.geometry.attributes.position.needsUpdate = true
}

const updateColor = () => {
    if (!deformedLine || !props.nodes.length) return

    const modeShape = combinedModeShape.value
    let maxDisp = 0
    props.nodes.forEach((_, idx) => {
        const dx = Math.abs(modeShape[idx * 6] || 0)
        const dy = Math.abs(modeShape[idx * 6 + 1] || 0)
        const dz = Math.abs(modeShape[idx * 6 + 2] || 0)
        maxDisp = Math.max(maxDisp, dx, dy, dz)
    })

    if (maxDisp > 0) {
        const colors = new Float32Array((deformedLine.geometry.attributes.position.count as number) * 3)
        let pointIndex = 0
        props.elements.forEach(elem => {
            const nodeIdx1 = props.nodes.findIndex(n => n.id === elem.node1)
            const nodeIdx2 = props.nodes.findIndex(n => n.id === elem.node2)

            const disp1 = Math.sqrt(
                Math.pow(modeShape[nodeIdx1 * 6] || 0, 2) +
                Math.pow(modeShape[nodeIdx1 * 6 + 1] || 0, 2) +
                Math.pow(modeShape[nodeIdx1 * 6 + 2] || 0, 2)
            )
            const disp2 = Math.sqrt(
                Math.pow(modeShape[nodeIdx2 * 6] || 0, 2) +
                Math.pow(modeShape[nodeIdx2 * 6 + 1] || 0, 2) +
                Math.pow(modeShape[nodeIdx2 * 6 + 2] || 0, 2)
            )

            const color1 = getColor(disp1 / maxDisp)
            const color2 = getColor(disp2 / maxDisp)

            colors[pointIndex * 3] = color1.r
            colors[pointIndex * 3 + 1] = color1.g
            colors[pointIndex * 3 + 2] = color1.b
            pointIndex++

            colors[pointIndex * 3] = color2.r
            colors[pointIndex * 3 + 1] = color2.g
            colors[pointIndex * 3 + 2] = color2.b
            pointIndex++
        })
        deformedLine.geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3))
        ;(deformedLine.material as THREE.LineBasicMaterial).vertexColors = true
    }
}

const getColor = (value: number) => {
    const r = Math.min(1, value * 2)
    const g = Math.min(1, (1 - value) * 2)
    const b = Math.max(0, 1 - value * 2)
    return { r, g, b }
}

const animate = () => {
    animationId = requestAnimationFrame(animate)

    if (isAnimating.value && combinedModeShape.value.length > 0) {
        time += 0.02 * animationSpeed.value
        const factor = Math.sin(time)
        updateDeformedShape(factor)
    }

    controls.update()
    renderer.render(scene, camera)
}

const toggleAnimation = () => {
    isAnimating.value = !isAnimating.value
}

watch(() => props.nodes, () => {
    createStructure()
    updateColor()
}, { deep: true })

watch(() => props.elements, () => {
    createStructure()
    updateColor()
}, { deep: true })

watch(combinedModeShape, () => {
    updateColor()
    if (!isAnimating.value) {
        updateDeformedShape(1)
    }
}, { deep: true })

watch(scaleFactor, () => {
    if (!isAnimating.value) {
        updateDeformedShape(1)
    }
})

const handleResize = () => {
    if (!container.value) return
    camera.aspect = container.value.clientWidth / container.value.clientHeight
    camera.updateProjectionMatrix()
    renderer.setSize(container.value.clientWidth, container.value.clientHeight)
}

onMounted(() => {
    initScene()
    window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
    cancelAnimationFrame(animationId)
    window.removeEventListener('resize', handleResize)
    renderer.dispose()
})
</script>

<style scoped>
.viewer-container {
    position: relative;
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
}

.canvas-container {
    flex: 1;
    position: relative;
}

.controls {
    position: absolute;
    top: 10px;
    right: 10px;
    background: rgba(255, 255, 255, 0.9);
    padding: 15px;
    border-radius: 8px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
    z-index: 100;
}

.control-group {
    margin-bottom: 15px;
}

.control-group:last-child {
    margin-bottom: 0;
}

.control-group label {
    display: block;
    margin-bottom: 5px;
    font-size: 12px;
    color: #666;
}

.control-group span {
    display: inline-block;
    margin-left: 10px;
    font-size: 12px;
    color: #666;
}

.color-bar {
    position: absolute;
    bottom: 10px;
    left: 10px;
    display: flex;
    align-items: center;
    background: rgba(255, 255, 255, 0.9);
    padding: 5px 10px;
    border-radius: 4px;
    font-size: 12px;
    color: #666;
}

.color-gradient {
    width: 100px;
    height: 15px;
    margin: 0 10px;
    border-radius: 3px;
    background: linear-gradient(to right, blue, cyan, green, yellow, red);
}
</style>
