export interface Node {
    id: number
    x: number
    y: number
    z: number
}

export interface Element {
    id: number
    node1: number
    node2: number
    sectionId: number
}

export interface Section {
    id: number
    a: number
    ix: number
    iy: number
    iz: number
    e: number
    rho: number
    nu: number
}

export interface Constraint {
    nodeId: number
    dx: boolean
    dy: boolean
    dz: boolean
    rx: boolean
    ry: boolean
    rz: boolean
}

export interface AnalysisRequest {
    nodes: Node[]
    elements: Element[]
    sections: Section[]
    constraints: Constraint[]
    numModes: number
}

export interface ModalResult {
    frequencyHz: number
    circularFreq: number
    period: number
    modeShape: number[]
    massParticipation: number[]
    isRigidBody: boolean
}

export interface AnalysisResponse {
    success: boolean
    message: string
    modalResults: ModalResult[]
    theoryValues?: TheoryValue[]
}

export interface TheoryValue {
    mode: number
    frequencyHz: number
    error: number
}

export interface Preset {
    name: string
    description: string
}

export interface FRFRequest {
    nodes: Node[]
    elements: Element[]
    sections: Section[]
    constraints: Constraint[]
    excitationNode: number
    direction: string
    amplitude: number
    freqStart: number
    freqEnd: number
    numPoints: number
    dampingRatios: number[]
}

export interface FRFResponse {
    success: boolean
    message: string
    frequencies: number[]
    amplitudes: number[]
    resonances: number[]
}

export interface TransientRequest {
    nodes: Node[]
    elements: Element[]
    sections: Section[]
    constraints: Constraint[]
    excitationNode: number
    direction: string
    waveformType: 'impulse' | 'step' | 'halfsine'
    amplitude: number
    duration: number
    timeStep: number
    totalTime: number
    dampingRatio: number
    observationNode: number
    observationDirection: string
}

export interface TransientResponse {
    success: boolean
    message: string
    timePoints: number[]
    displacements: number[]
    allDisplacements: number[][]
}

export interface ModeSelection {
    modeIndex: number
    enabled: boolean
    scale: number
}
