package handler

import (
    "math"
    "modal-analysis/internal/model"
    "modal-analysis/internal/preset"
    "modal-analysis/internal/solver"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
)

func AnalyzeModal(c *gin.Context) {
    var request model.AnalysisRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, model.AnalysisResponse{
            Success: false,
            Message: "Invalid request: " + err.Error(),
        })
        return
    }

    if len(request.Nodes) == 0 {
        c.JSON(http.StatusBadRequest, model.AnalysisResponse{
            Success: false,
            Message: "至少需要一个节点",
        })
        return
    }

    nodeIDs := make(map[int]bool)
    for _, node := range request.Nodes {
        if nodeIDs[node.ID] {
            c.JSON(http.StatusBadRequest, model.AnalysisResponse{
                Success: false,
                Message: "节点编号重复: " + strconv.Itoa(node.ID),
            })
            return
        }
        nodeIDs[node.ID] = true
    }

    for _, elem := range request.Elements {
        if !nodeIDs[elem.Node1] || !nodeIDs[elem.Node2] {
            c.JSON(http.StatusBadRequest, model.AnalysisResponse{
                Success: false,
                Message: "单元引用了不存在的节点: " + strconv.Itoa(elem.Node1) + " 或 " + strconv.Itoa(elem.Node2),
            })
            return
        }
    }

    sectionIDs := make(map[int]bool)
    for _, sec := range request.Sections {
        if sec.A <= 0 || sec.Ix <= 0 || sec.Iy <= 0 || sec.Iz <= 0 || sec.E <= 0 || sec.Rho <= 0 {
            c.JSON(http.StatusBadRequest, model.AnalysisResponse{
                Success: false,
                Message: "截面属性值必须为正数",
            })
            return
        }
        sectionIDs[sec.ID] = true
    }

    for _, elem := range request.Elements {
        if !sectionIDs[elem.SectionID] {
            c.JSON(http.StatusBadRequest, model.AnalysisResponse{
                Success: false,
                Message: "单元引用了不存在的截面属性: " + strconv.Itoa(elem.SectionID),
            })
            return
        }
    }

    if len(request.Constraints) == 0 {
        c.JSON(http.StatusBadRequest, model.AnalysisResponse{
            Success: false,
            Message: "至少需要一个约束条件",
        })
        return
    }

    K, M, _ := solver.AssembleGlobalMatrices(request.Nodes, request.Elements, request.Sections)
    
    _, err := solver.ApplyConstraints(K, M, request.Constraints, request.Nodes)
    if err != nil {
        c.JSON(http.StatusInternalServerError, model.AnalysisResponse{
            Success: false,
            Message: "应用约束失败: " + err.Error(),
        })
        return
    }

    if !K.IsPositiveDefinite() {
        c.JSON(http.StatusBadRequest, model.AnalysisResponse{
            Success: false,
            Message: "刚度矩阵不正定，结构可能是机构",
        })
        return
    }

    numModes := request.NumModes
    if numModes <= 0 {
        numModes = 10
    }
    if numModes > 50 {
        numModes = 50
    }

    eigenPairs, err := solver.SolveEigenvalue(K, M, numModes)
    if err != nil {
        c.JSON(http.StatusInternalServerError, model.AnalysisResponse{
            Success: false,
            Message: "特征值求解失败: " + err.Error(),
        })
        return
    }

    var modalResults []model.ModalResult
    var eigenvalues []float64
    var eigenvectors [][]float64

    for _, pair := range eigenPairs {
        freqHz := math.Sqrt(pair.Eigenvalue) / (2 * math.Pi)
        
        isRigidBody := false
        if freqHz < 1e-6 {
            isRigidBody = true
        }

        normalizedMode := solver.NormalizeByMass(pair.Eigenvector, M)
        participation := solver.CalculateMassParticipation(normalizedMode, M)

        modalResults = append(modalResults, model.ModalResult{
            FrequencyHz:       freqHz,
            CircularFreq:      math.Sqrt(pair.Eigenvalue),
            Period:            1 / freqHz,
            ModeShape:         normalizedMode,
            MassParticipation: participation,
            IsRigidBody:       isRigidBody,
        })

        eigenvalues = append(eigenvalues, pair.Eigenvalue)
        eigenvectors = append(eigenvectors, normalizedMode)
    }

    c.Set("eigenvalues", eigenvalues)
    c.Set("eigenvectors", eigenvectors)

    c.JSON(http.StatusOK, model.AnalysisResponse{
        Success:      true,
        Message:      "分析完成",
        ModalResults: modalResults,
    })
}

func GetPresets(c *gin.Context) {
    presets := preset.GetPresets()
    var result []gin.H
    for _, p := range presets {
        result = append(result, gin.H{
            "name":        p.Name,
            "description": p.Description,
        })
    }
    c.JSON(http.StatusOK, result)
}

func LoadPreset(c *gin.Context) {
    name := c.Param("name")
    p := preset.GetPresetByName(name)
    if p == nil {
        c.JSON(http.StatusNotFound, gin.H{
            "success": false,
            "message": "预设算例不存在",
        })
        return
    }

    K, M, _ := solver.AssembleGlobalMatrices(p.Model.Nodes, p.Model.Elements, p.Model.Sections)
    solver.ApplyConstraints(K, M, p.Model.Constraints, p.Model.Nodes)

    numModes := p.Model.NumModes
    eigenPairs, _ := solver.SolveEigenvalue(K, M, numModes)

    var modalResults []model.ModalResult
    for _, pair := range eigenPairs {
        freqHz := math.Sqrt(pair.Eigenvalue) / (2 * math.Pi)
        
        isRigidBody := false
        if freqHz < 1e-6 {
            isRigidBody = true
        }

        normalizedMode := solver.NormalizeByMass(pair.Eigenvector, M)
        participation := solver.CalculateMassParticipation(normalizedMode, M)

        modalResults = append(modalResults, model.ModalResult{
            FrequencyHz:       freqHz,
            CircularFreq:      math.Sqrt(pair.Eigenvalue),
            Period:            1 / freqHz,
            ModeShape:         normalizedMode,
            MassParticipation: participation,
            IsRigidBody:       isRigidBody,
        })
    }

    var theoryValues []model.TheoryValue
    for i, theoryFreq := range p.TheoryFreq {
        if i < len(modalResults) {
            errorPercent := math.Abs((modalResults[i].FrequencyHz - theoryFreq) / theoryFreq * 100)
            theoryValues = append(theoryValues, model.TheoryValue{
                Mode:       i + 1,
                FrequencyHz: theoryFreq,
                Error:      errorPercent,
            })
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "success":      true,
        "message":      "加载成功",
        "model":        p.Model,
        "modalResults": modalResults,
        "theoryValues": theoryValues,
    })
}

func CalculateFRF(c *gin.Context) {
    var request model.FRFRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, model.FRFResponse{
            Success: false,
            Message: "Invalid request: " + err.Error(),
        })
        return
    }

    K, M, _ := solver.AssembleGlobalMatrices(request.Nodes, request.Elements, request.Sections)
    solver.ApplyConstraints(K, M, request.Constraints, request.Nodes)

    numModes := 20
    eigenPairs, _ := solver.SolveEigenvalue(K, M, numModes)

    var eigenvalues []float64
    var eigenvectors [][]float64
    for _, pair := range eigenPairs {
        eigenvalues = append(eigenvalues, pair.Eigenvalue)
        normalizedMode := solver.NormalizeByMass(pair.Eigenvector, M)
        eigenvectors = append(eigenvectors, normalizedMode)
    }

    directionIndex := 0
    switch request.Direction {
    case "Y":
        directionIndex = 1
    case "Z":
        directionIndex = 2
    }

    nodeIndex := -1
    for i, node := range request.Nodes {
        if node.ID == request.ExcitationNode {
            nodeIndex = i
            break
        }
    }

    if nodeIndex == -1 {
        c.JSON(http.StatusBadRequest, model.FRFResponse{
            Success: false,
            Message: "激励节点不存在",
        })
        return
    }

    frequencies, amplitudes, resonances := solver.CalculateFRF(
        K, M, eigenvalues, eigenvectors,
        nodeIndex, directionIndex,
        request.Amplitude,
        request.FreqStart, request.FreqEnd,
        request.NumPoints,
        request.DampingRatios,
    )

    c.JSON(http.StatusOK, model.FRFResponse{
        Success:     true,
        Message:     "FRF计算完成",
        Frequencies: frequencies,
        Amplitudes:  amplitudes,
        Resonances:  resonances,
    })
}

func CalculateTransient(c *gin.Context) {
    var request model.TransientRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, model.TransientResponse{
            Success: false,
            Message: "Invalid request: " + err.Error(),
        })
        return
    }

    K, M, _ := solver.AssembleGlobalMatrices(request.Nodes, request.Elements, request.Sections)
    solver.ApplyConstraints(K, M, request.Constraints, request.Nodes)

    numModes := 20
    eigenPairs, _ := solver.SolveEigenvalue(K, M, numModes)

    var eigenvalues []float64
    var eigenvectors [][]float64
    for _, pair := range eigenPairs {
        eigenvalues = append(eigenvalues, pair.Eigenvalue)
        normalizedMode := solver.NormalizeByMass(pair.Eigenvector, M)
        eigenvectors = append(eigenvectors, normalizedMode)
    }

    directionIndex := 0
    switch request.Direction {
    case "Y":
        directionIndex = 1
    case "Z":
        directionIndex = 2
    }

    observationDirIndex := 0
    switch request.ObservationDirection {
    case "Y":
        observationDirIndex = 1
    case "Z":
        observationDirIndex = 2
    }

    excitationNodeIndex := -1
    for i, node := range request.Nodes {
        if node.ID == request.ExcitationNode {
            excitationNodeIndex = i
            break
        }
    }

    if excitationNodeIndex == -1 {
        c.JSON(http.StatusBadRequest, model.TransientResponse{
            Success: false,
            Message: "激励节点不存在",
        })
        return
    }

    observationNodeIndex := -1
    for i, node := range request.Nodes {
        if node.ID == request.ObservationNode {
            observationNodeIndex = i
            break
        }
    }

    if observationNodeIndex == -1 {
        c.JSON(http.StatusBadRequest, model.TransientResponse{
            Success: false,
            Message: "观测节点不存在",
        })
        return
    }

    timePoints, displacements, allDisplacements := solver.CalculateTransientResponse(
        K, M,
        eigenvalues,
        eigenvectors,
        excitationNodeIndex,
        directionIndex,
        request.WaveformType,
        request.Amplitude,
        request.Duration,
        request.TimeStep,
        request.TotalTime,
        request.DampingRatio,
        observationNodeIndex,
        observationDirIndex,
    )

    c.JSON(http.StatusOK, model.TransientResponse{
        Success:           true,
        Message:           "瞬态响应计算完成",
        TimePoints:        timePoints,
        Displacements:     displacements,
        AllDisplacements:  allDisplacements,
    })
}
