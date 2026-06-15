package solver

import (
    "math"
    "sort"
    "modal-analysis/internal/model"
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
        D[i] = 1.0 / math.Sqrt(M.At(i, i))
    }
    
    A := NewMatrix(n, n)
    for i := 0; i < n; i++ {
        for j := 0; j < n; j++ {
            A.Set(i, j, D[i]*K.At(i, j)*D[j])
        }
    }
    
    eigenvalues, eigenvectors := powerIteration(A, numModes)
    
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

func powerIteration(A *Matrix, numModes int) ([]float64, [][]float64) {
    n := A.Rows()
    eigenvalues := make([]float64, numModes)
    eigenvectors := make([][]float64, numModes)
    
    for m := 0; m < numModes; m++ {
        v := make([]float64, n)
        for i := range v {
            v[i] = 1.0 / math.Sqrt(float64(n))
        }
        
        for i := 0; i < 1000; i++ {
            vNew := A.VectorMultiply(v)
            
            norm := 0.0
            for _, val := range vNew {
                norm += val * val
            }
            norm = math.Sqrt(norm)
            
            for j := range v {
                v[j] = vNew[j] / norm
            }
        }
        
        lambda := 0.0
        Av := A.VectorMultiply(v)
        for i := 0; i < n; i++ {
            lambda += v[i] * Av[i]
        }
        
        eigenvalues[m] = lambda
        eigenvectors[m] = v
        
        for i := 0; i < n; i++ {
            for j := 0; j < n; j++ {
                A.Set(i, j, A.At(i, j)-lambda*v[i]*v[j])
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
    nodeIDs := make(map[int]bool)
    for _, node := range request.Nodes {
        if nodeIDs[node.ID] {
            return nil
        }
        nodeIDs[node.ID] = true
    }
    
    for _, elem := range request.Elements {
        if !nodeIDs[elem.Node1] || !nodeIDs[elem.Node2] {
            return nil
        }
    }
    
    sectionIDs := make(map[int]bool)
    for _, sec := range request.Sections {
        if sec.A <= 0 || sec.Ix <= 0 || sec.Iy <= 0 || sec.Iz <= 0 || sec.E <= 0 || sec.Rho <= 0 {
            return nil
        }
        sectionIDs[sec.ID] = true
    }
    
    for _, elem := range request.Elements {
        if !sectionIDs[elem.SectionID] {
            return nil
        }
    }
    
    if len(request.Constraints) == 0 {
        return nil
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
        sum := complex(0, 0)
        
        for j := 0; j < n; j++ {
            omega_j := math.Sqrt(eigenvalues[j])
            zeta_j := dampingRatios[j]
            
            phi_j := eigenvectors[j][dof]
            numerator := phi_j * phi_j * amplitude
            denominator := eigenvalues[j] - omega*omega + complex(0, 1)*2*zeta_j*omega_j*omega
            
            sum += complex(numerator, 0) / denominator
        }
        
        amplitudes[i] = math.Abs(complex128(sum))
        
        if i > 0 && i < numPoints-1 {
            if amplitudes[i] > amplitudes[i-1] && amplitudes[i] > amplitudes[i+1] {
                resonances = append(resonances, freq)
            }
        }
    }
    
    return frequencies, amplitudes, resonances
}
