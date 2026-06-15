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

export interface DesignVariable {
    sectionId: number
    property: 'a' | 'ix' | 'iy' | 'iz' | 'e' | 'rho'
    enabled: boolean
    lowerBound: number
    upperBound: number
    initialValue: number
}

export interface SensitivityResult {
    modeIndex: number
    frequencyHz: number
    sensitivities: {
        designVarId: string
        sensitivity: number
    }[]
}

export interface SensitivityResponse {
    success: boolean
    message: string
    results: SensitivityResult[]
    designVarNames: string[]
}

export interface FrequencyConstraint {
    modeIndex: number
    type: 'lower' | 'upper' | 'both'
    lowerBound?: number
    upperBound?: number
}

export interface OptimizationRequest {
    nodes: Node[]
    elements: Element[]
    sections: Section[]
    constraints: Constraint[]
    designVariables: DesignVariable[]
    frequencyConstraints: FrequencyConstraint[]
    numModes: number
    maxIterations: number
    tolerance: number
    constraintTolerance?: number
    designVarTolerance?: number
}

export interface IterationRecord {
    iteration: number
    objective: number
    designVariables: number[]
    frequencies: number[]
    constraintViolations: number[]
    feasible: boolean
}

export interface OptimizationResponse {
    success: boolean
    message: string
    initialDesignVariables: number[]
    finalDesignVariables: number[]
    initialFrequencies: number[]
    finalFrequencies: number[]
    initialMass: number
    finalMass: number
    iterations: IterationRecord[]
    converged: boolean
}

export interface ParamScanRequest {
    nodes: Node[]
    elements: Element[]
    sections: Section[]
    constraints: Constraint[]
    designVariable: DesignVariable
    scanStart: number
    scanEnd: number
    numSteps: number
    numModes: number
    secondVariable?: DesignVariable
    secondScanStart?: number
    secondScanEnd?: number
    secondNumSteps?: number
}

export interface ParamScanResponse {
    success: boolean
    message: string
    scanValues: number[]
    frequencies: number[][]
    secondScanValues?: number[]
    frequencies2D?: number[][]
}

export interface OptimizationHistoryRecord {
    id: string
    timestamp: string
    name: string
    initialDesignVariables: number[]
    finalDesignVariables: number[]
    initialFrequencies: number[]
    finalFrequencies: number[]
    initialMass: number
    finalMass: number
    iterations: IterationRecord[]
    converged: boolean
    designVarNames: string[]
}
