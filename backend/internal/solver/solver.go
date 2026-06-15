package solver

import (
    "errors"
    "math"
    "sort"
    "modal-analysis/internal/model"
    "strconv"
)

type EigenPair struct {
    Eigenvalue float64
    Eigenvector []float64
}

func AssembleGlobalMatrices(nodes []model.Node, elements []model.Element, sections []model.Section) (*Matrix, *Matrix, []int) {
    numNodes := len(nodes)
    totalDofs := numNodes * 6
    
    K := NewMatrix(totalDofs, totalDofs)
    M := NewMatrix(totalDofs, totalDofs)
    
    nodeIndex := make(map[int]int)
    for i, node := range nodes {
        nodeIndex[node.ID] = i
    }
    
    sectionIndex := make(map[int]model.Section)
    for _, sec := range sections {
        sectionIndex[sec.ID] = sec
    }
    
    for _, elem := range elements {
        idx1 := nodeIndex[elem.Node1]
        idx2 := nodeIndex[elem.Node2]
        
        node1 := nodes[idx1]
        node2 := nodes[idx2]
        
        dx := node2.X - node1.X
        dy := node2.Y - node1.Y
        dz := node2.Z - node1.Z
        length := math.Sqrt(dx*dx + dy*dy + dz*dz)
        
        section := sectionIndex[elem.SectionID]
        
        kLocal := ComputeLocalStiffnessMatrix(section, length)
        mLocal := ComputeLocalMassMatrix(section, length)
        T := ComputeTransformationMatrix(node1, node2)
        
        Tt := T.Transpose()
        kGlobal := Tt.Multiply(kLocal).Multiply(T)
        mGlobal := Tt.Multiply(mLocal).Multiply(T)
        
        dofs1 := []int{idx1*6, idx1*6+1, idx1*6+2, idx1*6+3, idx1*6+4, idx1*6+5}
        dofs2 := []int{idx2*6, idx2*6+1, idx2*6+2, idx2*6+3, idx2*6+4, idx2*6+5}
        elemDofs := append(dofs1, dofs2...)
        
        for i := 0; i < 12; i++ {
            for j := 0; j < 12; j++ {
                K.Set(elemDofs[i], elemDofs[j], K.At(elemDofs[i], elemDofs[j])+kGlobal.At(i, j))
                M.Set(elemDofs[i], elemDofs[j], M.At(elemDofs[i], elemDofs[j])+mGlobal.At(i, j))
            }
        }
    }
    
    return K, M, nil
}

func ApplyConstraints(K, M *Matrix, constraints []model.Constraint, nodes []model.Node) ([]int, error) {
    nodeIndex := make(map[int]int)
    for i, node := range nodes {
        nodeIndex[node.ID] = i
    }
    
    var fixedDofs []int
    
    for _, cons := range constraints {
        idx := nodeIndex[cons.NodeID]
        if cons.DX {
            fixedDofs = append(fixedDofs, idx*6)
        }
        if cons.DY {
            fixedDofs = append(fixedDofs, idx*6+1)
        }
        if cons.DZ {
            fixedDofs = append(fixedDofs, idx*6+2)
        }
        if cons.RX {
            fixedDofs = append(fixedDofs, idx*6+3)
        }
        if cons.RY {
            fixedDofs = append(fixedDofs, idx*6+4)
        }
        if cons.RZ {
            fixedDofs = append(fixedDofs, idx*6+5)
        }
    }
    
    K.ApplyBC(fixedDofs)
    M.ApplyBC(fixedDofs)
    
    return fixedDofs, nil
}

func SolveEigenvalue(K, M *Matrix, numModes int) ([]EigenPair, error) {
    n := K.Rows()
    
    D := make([]float64, n)
    for i := 0; i < n; i++ {
        mii := M.At(i, i)
        if mii < 1e-15 {
            D[i] = 0
        } else {
            D[i] = 1.0 / math.Sqrt(mii)
        }
    }
    
    A := NewMatrix(n, n)
    for i := 0; i < n; i++ {
        for j := 0; j < n; j++ {
            A.Set(i, j, D[i]*K.At(i, j)*D[j])
        }
    }
    
    eigenvalues, eigenvectors := jacobiMethod(A, numModes)
    
    var pairs []EigenPair
    for i := 0; i < len(eigenvalues); i++ {
        vec := make([]float64, n)
        for j := 0; j < n; j++ {
            vec[j] = eigenvectors[i][j] * D[j]
        }
        pairs = append(pairs, EigenPair{Eigenvalue: eigenvalues[i], Eigenvector: vec})
    }
    
    sort.Slice(pairs, func(i, j int) bool {
        return pairs[i].Eigenvalue < pairs[j].Eigenvalue
    })
    
    return pairs, nil
}

func jacobiMethod(A *Matrix, numModes int) ([]float64, [][]float64) {
    n := A.Rows()
    maxIter := 10000
    
    D := make([]float64, n)
    V := make([][]float64, n)
    for i := range V {
        V[i] = make([]float64, n)
        V[i][i] = 1
    }
    
    for i := 0; i < n; i++ {
        D[i] = A.At(i, i)
    }
    
    for iter := 0; iter < maxIter; iter++ {
        maxOff := 0.0
        p, q := 0, 1
        for i := 0; i < n; i++ {
            for j := i + 1; j < n; j++ {
                if math.Abs(A.At(i, j)) > maxOff {
                    maxOff = math.Abs(A.At(i, j))
                    p, q = i, j
                }
            }
        }
        
        if maxOff < 1e-12 {
            break
        }
        
        apq := A.At(p, q)
        app := A.At(p, p)
        aqq := A.At(q, q)
        
        theta := 0.5 * math.Atan2(2*apq, aqq-app)
        c := math.Cos(theta)
        s := math.Sin(theta)
        
        for i := 0; i < n; i++ {
            if i != p && i != q {
                api := A.At(p, i)
                aqi := A.At(q, i)
                A.Set(p, i, c*api-s*aqi)
                A.Set(i, p, A.At(p, i))
                A.Set(q, i, s*api+c*aqi)
                A.Set(i, q, A.At(q, i))
            }
        }
        
        app_new := c*c*app - 2*s*c*apq + s*s*aqq
        aqq_new := s*s*app + 2*s*c*apq + c*c*aqq
        A.Set(p, p, app_new)
        A.Set(q, q, aqq_new)
        A.Set(p, q, 0)
        A.Set(q, p, 0)
        
        for i := 0; i < n; i++ {
            vpi := V[p][i]
            vqi := V[q][i]
            V[p][i] = c*vpi - s*vqi
            V[q][i] = s*vpi + c*vqi
        }
        
        D[p] = app_new
        D[q] = aqq_new
    }
    
    indices := make([]int, n)
    for i := range indices {
        indices[i] = i
    }
    
    sort.Slice(indices, func(i, j int) bool {
        return D[indices[i]] < D[indices[j]]
    })
    
    numModes = min(numModes, n)
    eigenvalues := make([]float64, numModes)
    eigenvectors := make([][]float64, numModes)
    
    for i := 0; i < numModes; i++ {
        eigenvalues[i] = D[indices[i]]
        eigenvectors[i] = make([]float64, n)
        for j := 0; j < n; j++ {
            eigenvectors[i][j] = V[indices[i]][j]
        }
        
        norm := 0.0
        for j := 0; j < n; j++ {
            norm += eigenvectors[i][j] * eigenvectors[i][j]
        }
        norm = math.Sqrt(norm)
        if norm > 1e-15 {
            for j := 0; j < n; j++ {
                eigenvectors[i][j] /= norm
            }
        }
    }
    
    return eigenvalues, eigenvectors
}

func NormalizeByMass(modeShape []float64, M *Matrix) []float64 {
    n := len(modeShape)
    norm := 0.0
    
    for i := 0; i < n; i++ {
        for j := 0; j < n; j++ {
            norm += modeShape[i] * M.At(i, j) * modeShape[j]
        }
    }
    
    if norm < 1e-15 {
        return modeShape
    }
    
    norm = math.Sqrt(norm)
    result := make([]float64, n)
    for i := range result {
        result[i] = modeShape[i] / norm
    }
    
    return result
}

func CalculateMassParticipation(modeShape []float64, M *Matrix) []float64 {
    n := len(modeShape)
    totalMass := 0.0
    for i := 0; i < n; i += 6 {
        for j := i; j < i+3; j++ {
            for k := i; k < i+3; k++ {
                totalMass += M.At(j, k)
            }
        }
    }
    
    if totalMass < 1e-15 {
        return []float64{0, 0, 0}
    }
    
    participation := make([]float64, 3)
    for dir := 0; dir < 3; dir++ {
        numerator := 0.0
        for i := 0; i < n; i += 6 {
            for j := 0; j < n; j += 6 {
                numerator += modeShape[i+dir] * M.At(i+dir, j+dir)
            }
        }
        participation[dir] = numerator * numerator / totalMass
    }
    
    return participation
}

func ValidateModel(request model.AnalysisRequest) error {
    if len(request.Nodes) == 0 {
        return errors.New("至少需要一个节点")
    }
    
    nodeIDs := make(map[int]bool)
    for _, node := range request.Nodes {
        if nodeIDs[node.ID] {
            return errors.New("节点编号重复: " + strconv.Itoa(node.ID))
        }
        nodeIDs[node.ID] = true
    }
    
    if len(request.Elements) == 0 {
        return errors.New("至少需要一个单元")
    }
    
    for _, elem := range request.Elements {
        if !nodeIDs[elem.Node1] {
            return errors.New("单元引用了不存在的节点: " + strconv.Itoa(elem.Node1))
        }
        if !nodeIDs[elem.Node2] {
            return errors.New("单元引用了不存在的节点: " + strconv.Itoa(elem.Node2))
        }
    }
    
    if len(request.Sections) == 0 {
        return errors.New("至少需要一个截面属性")
    }
    
    sectionIDs := make(map[int]bool)
    for _, sec := range request.Sections {
        if sec.A <= 0 {
            return errors.New("截面属性的面积A必须大于0")
        }
        if sec.Ix <= 0 {
            return errors.New("截面属性的惯性矩Ix必须大于0")
        }
        if sec.Iy <= 0 {
            return errors.New("截面属性的惯性矩Iy必须大于0")
        }
        if sec.Iz <= 0 {
            return errors.New("截面属性的惯性矩Iz必须大于0")
        }
        if sec.E <= 0 {
            return errors.New("截面属性的弹性模量E必须大于0")
        }
        if sec.Rho <= 0 {
            return errors.New("截面属性的密度rho必须大于0")
        }
        if sec.Nu < 0 || sec.Nu > 0.5 {
            return errors.New("截面属性的泊松比nu必须在0到0.5之间")
        }
        sectionIDs[sec.ID] = true
    }
    
    for _, elem := range request.Elements {
        if !sectionIDs[elem.SectionID] {
            return errors.New("单元引用了不存在的截面属性: " + strconv.Itoa(elem.SectionID))
        }
    }
    
    if len(request.Constraints) == 0 {
        return errors.New("至少需要一个约束条件")
    }
    
    return nil
}

func CalculateFRF(K, M *Matrix, eigenvalues []float64, eigenvectors [][]float64, 
    excitationNode, directionIndex int, amplitude float64, 
    freqStart, freqEnd float64, numPoints int, dampingRatios []float64) ([]float64, []float64, []float64) {
    
    frequencies := make([]float64, numPoints)
    amplitudes := make([]float64, numPoints)
    
    df := (freqEnd - freqStart) / float64(numPoints-1)
    for i := 0; i < numPoints; i++ {
        frequencies[i] = freqStart + df*float64(i)
    }
    
    n := len(eigenvalues)
    if len(dampingRatios) < n {
        temp := make([]float64, n)
        copy(temp, dampingRatios)
        for i := len(dampingRatios); i < n; i++ {
            temp[i] = 0.02
        }
        dampingRatios = temp
    }
    
    dof := excitationNode * 6 + directionIndex
    
    var resonances []float64
    
    for i, freq := range frequencies {
        omega := 2 * math.Pi * freq
        var sumReal, sumImag float64
        
        for j := 0; j < n; j++ {
            omega_j := math.Sqrt(eigenvalues[j])
            zeta_j := dampingRatios[j]
            
            phi_j := eigenvectors[j][dof]
            num := phi_j * phi_j * amplitude
            
            denomReal := eigenvalues[j] - omega*omega
            denomImag := 2 * zeta_j * omega_j * omega
            denomMag2 := denomReal*denomReal + denomImag*denomImag
            
            if denomMag2 > 1e-30 {
                sumReal += num * denomReal / denomMag2
                sumImag += num * denomImag / denomMag2
            }
        }
        
        amplitudes[i] = math.Sqrt(sumReal*sumReal + sumImag*sumImag)
        
        if i > 0 && i < numPoints-1 {
            if amplitudes[i] > amplitudes[i-1] && amplitudes[i] > amplitudes[i+1] {
                resonances = append(resonances, freq)
            }
        }
    }
    
    return frequencies, amplitudes, resonances
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func GetExcitationForce(waveformType string, amplitude, duration, time float64) float64 {
    switch waveformType {
    case "impulse":
        if time < duration {
            return amplitude / duration
        }
        return 0
    case "step":
        if time >= 0 {
            return amplitude
        }
        return 0
    case "halfsine":
        if time >= 0 && time < duration {
            return amplitude * math.Sin(math.Pi * time / duration)
        }
        return 0
    default:
        return 0
    }
}

func CalculateTransientResponse(
    K, M *Matrix,
    eigenvalues []float64,
    eigenvectors [][]float64,
    excitationNode, directionIndex int,
    waveformType string,
    amplitude, duration, timeStep, totalTime, dampingRatio float64,
    observationNode, observationDirIndex int) ([]float64, []float64, [][]float64) {

    numModes := len(eigenvalues)
    numTimeSteps := int(totalTime / timeStep)
    if numTimeSteps <= 0 {
        numTimeSteps = 1000
    }

    timePoints := make([]float64, numTimeSteps)
    for i := 0; i < numTimeSteps; i++ {
        timePoints[i] = float64(i) * timeStep
    }

    excitationDof := excitationNode*6 + directionIndex
    observationDof := observationNode*6 + observationDirIndex

    allDisplacements := make([][]float64, numTimeSteps)
    for i := range allDisplacements {
        allDisplacements[i] = make([]float64, len(eigenvectors[0]))
    }

    displacements := make([]float64, numTimeSteps)

    for i, t := range timePoints {
        f := GetExcitationForce(waveformType, amplitude, duration, t)
        
        var totalDisp float64
        
        for j := 0; j < numModes; j++ {
            omega_j := math.Sqrt(eigenvalues[j])
            zeta_j := dampingRatio
            
            phi_j_exc := eigenvectors[j][excitationDof]
            phi_j_obs := eigenvectors[j][observationDof]
            
            if phi_j_exc == 0 {
                continue
            }
            
            omega_d := omega_j * math.Sqrt(1-zeta_j*zeta_j)
            
            var disp float64
            
            if t >= duration {
                tau := t - duration
                disp = (f * phi_j_exc / (omega_d * eigenvalues[j])) * 
                    math.Exp(-zeta_j*omega_j*t) * 
                    math.Sin(omega_d*t) -
                    (f * phi_j_exc / (omega_d * eigenvalues[j])) * 
                    math.Exp(-zeta_j*omega_j*tau) * 
                    math.Sin(omega_d*tau)
            } else {
                switch waveformType {
                case "impulse":
                    disp = (f * phi_j_exc / eigenvalues[j]) * 
                        math.Exp(-zeta_j*omega_j*t) * 
                        math.Sin(omega_d*t) / omega_d
                case "step":
                    disp = (f * phi_j_exc / eigenvalues[j]) * 
                        (1 - math.Exp(-zeta_j*omega_j*t) * 
                        (math.Cos(omega_d*t) + zeta_j*omega_j*math.Sin(omega_d*t)/omega_d))
                case "halfsine":
                    wd_2 := omega_d * omega_d
                    wn_2 := eigenvalues[j]
                    zeta_wn := zeta_j * omega_j
                    
                    A := f * phi_j_exc * duration / (math.Pi * (wn_2 - (math.Pi/duration)*(math.Pi/duration)))
                    
                    term1 := math.Exp(-zeta_wn*t) * 
                        (math.Cos(omega_d*t) + zeta_wn*math.Sin(omega_d*t)/omega_d)
                    
                    term2 := math.Cos(math.Pi * t / duration)
                    term3 := (2 * zeta_wn * math.Pi / duration) * math.Sin(math.Pi * t / duration) / wd_2
                    term4 := ((wn_2 - (math.Pi/duration)*(math.Pi/duration)) / wd_2) * math.Sin(math.Pi * t / duration)
                    
                    disp = A * (term1 - term2 + term3 + term4)
                }
            }
            
            totalDisp += disp * phi_j_obs
            
            for k := range allDisplacements[i] {
                allDisplacements[i][k] += disp * eigenvectors[j][k]
            }
        }
        
        displacements[i] = totalDisp
    }

    return timePoints, displacements, allDisplacements
}
