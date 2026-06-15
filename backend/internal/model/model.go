package model

type Node struct {
    ID     int     `json:"id" validate:"required,min=1"`
    X      float64 `json:"x" validate:"required"`
    Y      float64 `json:"y" validate:"required"`
    Z      float64 `json:"z" validate:"required"`
}

type Element struct {
    ID         int `json:"id" validate:"required,min=1"`
    Node1      int `json:"node1" validate:"required,min=1"`
    Node2      int `json:"node2" validate:"required,min=1"`
    SectionID  int `json:"sectionId" validate:"required,min=1"`
}

type Section struct {
    ID     int     `json:"id" validate:"required,min=1"`
    A      float64 `json:"a" validate:"required,gt=0"`
    Ix     float64 `json:"ix" validate:"required,gt=0"`
    Iy     float64 `json:"iy" validate:"required,gt=0"`
    Iz     float64 `json:"iz" validate:"required,gt=0"`
    E      float64 `json:"e" validate:"required,gt=0"`
    Rho    float64 `json:"rho" validate:"required,gt=0"`
    Nu     float64 `json:"nu" validate:"required,gte=0,lte=0.5"`
}

type Constraint struct {
    NodeID     int     `json:"nodeId" validate:"required,min=1"`
    DX         bool    `json:"dx"`
    DY         bool    `json:"dy"`
    DZ         bool    `json:"dz"`
    RX         bool    `json:"rx"`
    RY         bool    `json:"ry"`
    RZ         bool    `json:"rz"`
}

type AnalysisRequest struct {
    Nodes       []Node       `json:"nodes" validate:"required"`
    Elements    []Element    `json:"elements" validate:"required"`
    Sections    []Section    `json:"sections" validate:"required"`
    Constraints []Constraint `json:"constraints" validate:"required"`
    NumModes    int          `json:"numModes" validate:"min=1,max=50"`
}

type ModalResult struct {
    FrequencyHz    float64   `json:"frequencyHz"`
    CircularFreq   float64   `json:"circularFreq"`
    Period         float64   `json:"period"`
    ModeShape      []float64 `json:"modeShape"`
    MassParticipation []float64 `json:"massParticipation"`
    IsRigidBody    bool      `json:"isRigidBody"`
}

type AnalysisResponse struct {
    Success      bool          `json:"success"`
    Message      string        `json:"message"`
    ModalResults []ModalResult `json:"modalResults"`
    TheoryValues []TheoryValue `json:"theoryValues"`
}

type TheoryValue struct {
    Mode       int     `json:"mode"`
    FrequencyHz float64 `json:"frequencyHz"`
    Error      float64 `json:"error"`
}

type FRFRequest struct {
    AnalysisRequest
    ExcitationNode int     `json:"excitationNode" validate:"required,min=1"`
    Direction      string  `json:"direction" validate:"required,oneof=X Y Z"`
    Amplitude      float64 `json:"amplitude" validate:"required"`
    FreqStart      float64 `json:"freqStart" validate:"required,gte=0"`
    FreqEnd        float64 `json:"freqEnd" validate:"required,gt=freqStart"`
    NumPoints      int     `json:"numPoints" validate:"required,min=10"`
    DampingRatios  []float64 `json:"dampingRatios"`
}

type FRFResponse struct {
    Success     bool        `json:"success"`
    Message     string      `json:"message"`
    Frequencies []float64   `json:"frequencies"`
    Amplitudes  []float64   `json:"amplitudes"`
    Resonances  []float64   `json:"resonances"`
}
