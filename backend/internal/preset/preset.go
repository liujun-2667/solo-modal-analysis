package preset

import (
    "math"
    "modal-analysis/internal/model"
)

type Preset struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Model       model.AnalysisRequest `json:"model"`
    TheoryFreq  []float64       `json:"theoryFreq"`
}

var presets []Preset

func LoadPresets() {
    presets = []Preset{
        createSimplySupportedBeam(),
        createCantileverBeam(),
        createPortalFrame(),
        createThreeSpanBeam(),
        createSpaceTruss(),
    }
}

func GetPresets() []Preset {
    return presets
}

func GetPresetByName(name string) *Preset {
    for i := range presets {
        if presets[i].Name == name {
            return &presets[i]
        }
    }
    return nil
}

func createSimplySupportedBeam() Preset {
    L := 10.0
    E := 2.1e11
    rho := 7850.0
    A := 0.01
    Ix := 8.333e-6
    
    nodes := []model.Node{
        {ID: 1, X: 0, Y: 0, Z: 0},
        {ID: 2, X: 2.5, Y: 0, Z: 0},
        {ID: 3, X: 5, Y: 0, Z: 0},
        {ID: 4, X: 7.5, Y: 0, Z: 0},
        {ID: 5, X: 10, Y: 0, Z: 0},
    }
    
    elements := []model.Element{
        {ID: 1, Node1: 1, Node2: 2, SectionID: 1},
        {ID: 2, Node1: 2, Node2: 3, SectionID: 1},
        {ID: 3, Node1: 3, Node2: 4, SectionID: 1},
        {ID: 4, Node1: 4, Node2: 5, SectionID: 1},
    }
    
    sections := []model.Section{
        {ID: 1, A: A, Ix: Ix, Iy: 1e-6, Iz: 1e-6, E: E, Rho: rho, Nu: 0.3},
    }
    
    constraints := []model.Constraint{
        {NodeID: 1, DX: true, DY: true, DZ: true, RX: false, RY: false, RZ: false},
        {NodeID: 5, DX: true, DY: true, DZ: true, RX: false, RY: false, RZ: false},
    }
    
    theoryFreq := make([]float64, 3)
    for i := 1; i <= 3; i++ {
        theoryFreq[i-1] = float64(i*i) * math.Pi * math.Pi * math.Sqrt(E*Ix/(rho*A*L*L*L*L)) / (2 * math.Pi)
    }
    
    return Preset{
        Name:        "simply_supported_beam",
        Description: "简支梁 - 两端铰支",
        Model: model.AnalysisRequest{
            Nodes:       nodes,
            Elements:    elements,
            Sections:    sections,
            Constraints: constraints,
            NumModes:    10,
        },
        TheoryFreq: theoryFreq,
    }
}

func createCantileverBeam() Preset {
    L := 10.0
    E := 2.1e11
    rho := 7850.0
    A := 0.01
    Ix := 8.333e-6
    
    nodes := []model.Node{
        {ID: 1, X: 0, Y: 0, Z: 0},
        {ID: 2, X: 2.5, Y: 0, Z: 0},
        {ID: 3, X: 5, Y: 0, Z: 0},
        {ID: 4, X: 7.5, Y: 0, Z: 0},
        {ID: 5, X: 10, Y: 0, Z: 0},
    }
    
    elements := []model.Element{
        {ID: 1, Node1: 1, Node2: 2, SectionID: 1},
        {ID: 2, Node1: 2, Node2: 3, SectionID: 1},
        {ID: 3, Node1: 3, Node2: 4, SectionID: 1},
        {ID: 4, Node1: 4, Node2: 5, SectionID: 1},
    }
    
    sections := []model.Section{
        {ID: 1, A: A, Ix: Ix, Iy: 1e-6, Iz: 1e-6, E: E, Rho: rho, Nu: 0.3},
    }
    
    constraints := []model.Constraint{
        {NodeID: 1, DX: true, DY: true, DZ: true, RX: true, RY: true, RZ: true},
    }
    
    beta := []float64{1.8751, 4.6941, 7.8548}
    theoryFreq := make([]float64, 3)
    for i := 0; i < 3; i++ {
        theoryFreq[i] = beta[i] * beta[i] * math.Sqrt(E*Ix/(rho*A*L*L*L*L)) / (2 * math.Pi)
    }
    
    return Preset{
        Name:        "cantilever_beam",
        Description: "悬臂梁 - 一端固支",
        Model: model.AnalysisRequest{
            Nodes:       nodes,
            Elements:    elements,
            Sections:    sections,
            Constraints: constraints,
            NumModes:    10,
        },
        TheoryFreq: theoryFreq,
    }
}

func createPortalFrame() Preset {
    E := 2.1e11
    rho := 7850.0
    
    nodes := []model.Node{
        {ID: 1, X: 0, Y: 0, Z: 0},
        {ID: 2, X: 0, Y: 5, Z: 0},
        {ID: 3, X: 8, Y: 5, Z: 0},
        {ID: 4, X: 8, Y: 0, Z: 0},
    }
    
    elements := []model.Element{
        {ID: 1, Node1: 1, Node2: 2, SectionID: 1},
        {ID: 2, Node1: 2, Node2: 3, SectionID: 2},
        {ID: 3, Node1: 3, Node2: 4, SectionID: 1},
    }
    
    sections := []model.Section{
        {ID: 1, A: 0.01, Ix: 8.333e-6, Iy: 1e-6, Iz: 8.333e-6, E: E, Rho: rho, Nu: 0.3},
        {ID: 2, A: 0.015, Ix: 1.5e-5, Iy: 1e-6, Iz: 1.5e-5, E: E, Rho: rho, Nu: 0.3},
    }
    
    constraints := []model.Constraint{
        {NodeID: 1, DX: true, DY: true, DZ: true, RX: true, RY: true, RZ: true},
        {NodeID: 4, DX: true, DY: true, DZ: true, RX: true, RY: true, RZ: true},
    }
    
    return Preset{
        Name:        "portal_frame",
        Description: "门型框架 - 两柱一梁",
        Model: model.AnalysisRequest{
            Nodes:       nodes,
            Elements:    elements,
            Sections:    sections,
            Constraints: constraints,
            NumModes:    10,
        },
        TheoryFreq: []float64{},
    }
}

func createThreeSpanBeam() Preset {
    E := 2.1e11
    rho := 7850.0
    A := 0.01
    Ix := 8.333e-6
    
    nodes := []model.Node{
        {ID: 1, X: 0, Y: 0, Z: 0},
        {ID: 2, X: 10, Y: 0, Z: 0},
        {ID: 3, X: 20, Y: 0, Z: 0},
        {ID: 4, X: 30, Y: 0, Z: 0},
    }
    
    elements := []model.Element{
        {ID: 1, Node1: 1, Node2: 2, SectionID: 1},
        {ID: 2, Node1: 2, Node2: 3, SectionID: 1},
        {ID: 3, Node1: 3, Node2: 4, SectionID: 1},
    }
    
    sections := []model.Section{
        {ID: 1, A: A, Ix: Ix, Iy: 1e-6, Iz: 1e-6, E: E, Rho: rho, Nu: 0.3},
    }
    
    constraints := []model.Constraint{
        {NodeID: 1, DX: true, DY: true, DZ: true, RX: false, RY: false, RZ: false},
        {NodeID: 2, DX: true, DY: true, DZ: true, RX: false, RY: false, RZ: false},
        {NodeID: 3, DX: true, DY: true, DZ: true, RX: false, RY: false, RZ: false},
        {NodeID: 4, DX: true, DY: true, DZ: true, RX: false, RY: false, RZ: false},
    }
    
    return Preset{
        Name:        "three_span_beam",
        Description: "三跨连续梁",
        Model: model.AnalysisRequest{
            Nodes:       nodes,
            Elements:    elements,
            Sections:    sections,
            Constraints: constraints,
            NumModes:    10,
        },
        TheoryFreq: []float64{},
    }
}

func createSpaceTruss() Preset {
    E := 2.1e11
    rho := 7850.0
    
    nodes := []model.Node{
        {ID: 1, X: 0, Y: 0, Z: 0},
        {ID: 2, X: 5, Y: 0, Z: 0},
        {ID: 3, X: 2.5, Y: 4.33, Z: 0},
        {ID: 4, X: 2.5, Y: 1.44, Z: 3.46},
    }
    
    elements := []model.Element{
        {ID: 1, Node1: 1, Node2: 2, SectionID: 1},
        {ID: 2, Node1: 2, Node2: 3, SectionID: 1},
        {ID: 3, Node1: 3, Node2: 1, SectionID: 1},
        {ID: 4, Node1: 1, Node2: 4, SectionID: 1},
        {ID: 5, Node1: 2, Node2: 4, SectionID: 1},
        {ID: 6, Node1: 3, Node2: 4, SectionID: 1},
    }
    
    sections := []model.Section{
        {ID: 1, A: 0.001, Ix: 1e-8, Iy: 1e-8, Iz: 1e-8, E: E, Rho: rho, Nu: 0.3},
    }
    
    constraints := []model.Constraint{
        {NodeID: 1, DX: true, DY: true, DZ: true, RX: true, RY: true, RZ: true},
        {NodeID: 2, DX: true, DY: true, DZ: true, RX: true, RY: true, RZ: true},
        {NodeID: 3, DX: true, DY: true, DZ: true, RX: true, RY: true, RZ: true},
    }
    
    return Preset{
        Name:        "space_truss",
        Description: "空间桁架",
        Model: model.AnalysisRequest{
            Nodes:       nodes,
            Elements:    elements,
            Sections:    sections,
            Constraints: constraints,
            NumModes:    10,
        },
        TheoryFreq: []float64{},
    }
}
