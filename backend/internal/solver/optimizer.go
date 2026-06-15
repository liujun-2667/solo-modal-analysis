package solver

import (
	"math"
	"modal-analysis/internal/model"
)

func CalculateSensitivity(request model.OptimizationRequest) (*model.SensitivityResponse, error) {
	enabledDVs := make([]model.DesignVariable, 0)
	designVarNames := make([]string, 0)
	
	for _, dv := range request.DesignVariables {
		if dv.Enabled {
			enabledDVs = append(enabledDVs, dv)
			designVarNames = append(designVarNames, designVarName(dv))
		}
	}
	
	if len(enabledDVs) == 0 {
		return &model.SensitivityResponse{
			Success:       false,
			Message:       "没有启用的设计变量",
			Results:       nil,
			DesignVarNames: designVarNames,
		}, nil
	}
	
	baseFreqs, err := computeFrequencies(request.Nodes, request.Elements, request.Sections, request.Constraints, request.NumModes)
	if err != nil {
		return nil, err
	}
	
	results := make([]model.SensitivityResult, len(baseFreqs))
	
	for i, freq := range baseFreqs {
		sensitivities := make([]model.SensitivityItem, len(enabledDVs))
		
		for j, dv := range enabledDVs {
			perturbedSections := perturbSection(request.Sections, dv, 0.01)
			perturbedFreqs, _ := computeFrequencies(request.Nodes, request.Elements, perturbedSections, request.Constraints, request.NumModes)
			
			perturbedSectionsNeg := perturbSection(request.Sections, dv, -0.01)
			perturbedFreqsNeg, _ := computeFrequencies(request.Nodes, request.Elements, perturbedSectionsNeg, request.Constraints, request.NumModes)
			
			if i < len(perturbedFreqs) && i < len(perturbedFreqsNeg) {
				h := dv.InitialValue * 0.01
				sensitivity := (perturbedFreqs[i] - perturbedFreqsNeg[i]) / (2 * h)
				sensitivities[j] = model.SensitivityItem{
					DesignVarId: designVarName(dv),
					Sensitivity: sensitivity,
				}
			} else {
				sensitivities[j] = model.SensitivityItem{
					DesignVarId: designVarName(dv),
					Sensitivity: 0,
				}
			}
		}
		
		results[i] = model.SensitivityResult{
			ModeIndex:    i + 1,
			FrequencyHz:  freq,
			Sensitivities: sensitivities,
		}
	}
	
	return &model.SensitivityResponse{
		Success:       true,
		Message:       "灵敏度分析完成",
		Results:       results,
		DesignVarNames: designVarNames,
	}, nil
}

func designVarName(dv model.DesignVariable) string {
	propNames := map[string]string{
		"a":    "A",
		"ix":   "Ix",
		"iy":   "Iy",
		"iz":   "Iz",
		"e":    "E",
		"rho":  "rho",
	}
	return "截面" + string(rune('0'+dv.SectionID)) + "." + propNames[dv.Property]
}

func perturbSection(sections []model.Section, dv model.DesignVariable, factor float64) []model.Section {
	result := make([]model.Section, len(sections))
	for i, sec := range sections {
		result[i] = sec
		if sec.ID == dv.SectionID {
			delta := dv.InitialValue * factor
			switch dv.Property {
			case "a":
				result[i].A += delta
			case "ix":
				result[i].Ix += delta
			case "iy":
				result[i].Iy += delta
			case "iz":
				result[i].Iz += delta
			case "e":
				result[i].E += delta
			case "rho":
				result[i].Rho += delta
			}
		}
	}
	return result
}

func computeFrequencies(nodes []model.Node, elements []model.Element, sections []model.Section, 
	constraints []model.Constraint, numModes int) ([]float64, error) {
	K, M, _ := AssembleGlobalMatrices(nodes, elements, sections)
	ApplyConstraints(K, M, constraints, nodes)
	
	eigenPairs, err := SolveEigenvalue(K, M, numModes)
	if err != nil {
		return nil, err
	}
	
	freqs := make([]float64, len(eigenPairs))
	for i, pair := range eigenPairs {
		freqs[i] = math.Sqrt(pair.Eigenvalue) / (2 * math.Pi)
	}
	return freqs, nil
}

func CalculateTotalMass(elements []model.Element, sections []model.Section) float64 {
	sectionIndex := make(map[int]model.Section)
	for _, sec := range sections {
		sectionIndex[sec.ID] = sec
	}
	
	totalMass := 0.0
	for _, elem := range elements {
		sec := sectionIndex[elem.SectionID]
		totalMass += sec.A * sec.Rho
	}
	return totalMass
}

func SolveOptimization(request model.OptimizationRequest) (*model.OptimizationResponse, error) {
	enabledDVs := make([]model.DesignVariable, 0)
	for _, dv := range request.DesignVariables {
		if dv.Enabled {
			enabledDVs = append(enabledDVs, dv)
		}
	}
	
	if len(enabledDVs) == 0 {
		return &model.OptimizationResponse{
			Success: false,
			Message: "没有启用的设计变量",
		}, nil
	}
	
	x := make([]float64, len(enabledDVs))
	for i, dv := range enabledDVs {
		x[i] = dv.InitialValue
	}
	
	lb := make([]float64, len(enabledDVs))
	ub := make([]float64, len(enabledDVs))
	for i, dv := range enabledDVs {
		lb[i] = dv.LowerBound
		ub[i] = dv.UpperBound
	}
	
	initialFreqs, err := computeFrequencies(request.Nodes, request.Elements, request.Sections, request.Constraints, request.NumModes)
	if err != nil {
		return nil, err
	}
	
	initialMass := calculateMass(x, request.Elements, request.Sections, enabledDVs)
	
	iterations := make([]model.IterationRecord, 0)
	
	for iter := 0; iter < request.MaxIterations; iter++ {
		currentSections := updateSections(request.Sections, enabledDVs, x)
		freqs, _ := computeFrequencies(request.Nodes, request.Elements, currentSections, request.Constraints, request.NumModes)
		
		mass := calculateMass(x, request.Elements, request.Sections, enabledDVs)
		
		constraintViolations := evaluateConstraints(freqs, request.FrequencyConstraints)
		feasible := true
		for _, v := range constraintViolations {
			if v > 0 {
				feasible = false
				break
			}
		}
		
		iterations = append(iterations, model.IterationRecord{
			Iteration:           iter + 1,
			Objective:          mass,
			DesignVariables:     append([]float64(nil), x...),
			Frequencies:        append([]float64(nil), freqs...),
			ConstraintViolations: append([]float64(nil), constraintViolations...),
			Feasible:           feasible,
		})
		
		if iter > 0 {
			if math.Abs(mass - iterations[iter-1].Objective) < request.Tolerance {
				break
			}
		}
		
		grad := computeGradient(x, request, enabledDVs)
		constraintGrads := computeConstraintGradients(x, request, enabledDVs)
		
		x = updateDesignVariables(x, grad, constraintGrads, lb, ub, request.FrequencyConstraints, freqs)
		
		for i := range x {
			if x[i] < lb[i] {
				x[i] = lb[i]
			}
			if x[i] > ub[i] {
				x[i] = ub[i]
			}
		}
	}
	
	finalSections := updateSections(request.Sections, enabledDVs, x)
	finalFreqs, _ := computeFrequencies(request.Nodes, request.Elements, finalSections, request.Constraints, request.NumModes)
	finalMass := calculateMass(x, request.Elements, request.Sections, enabledDVs)
	
	converged := len(iterations) > 1 && 
		math.Abs(finalMass - iterations[len(iterations)-2].Objective) < request.Tolerance
	
	return &model.OptimizationResponse{
		Success:              true,
		Message:              "优化完成",
		InitialDesignVariables: getInitialValues(enabledDVs),
		FinalDesignVariables:   append([]float64(nil), x...),
		InitialFrequencies:    initialFreqs,
		FinalFrequencies:      finalFreqs,
		InitialMass:          initialMass,
		FinalMass:            finalMass,
		Iterations:           iterations,
		Converged:            converged,
	}, nil
}

func getInitialValues(dvs []model.DesignVariable) []float64 {
	result := make([]float64, len(dvs))
	for i, dv := range dvs {
		result[i] = dv.InitialValue
	}
	return result
}

func updateSections(sections []model.Section, dvs []model.DesignVariable, x []float64) []model.Section {
	result := make([]model.Section, len(sections))
	copy(result, sections)
	
	dvIndex := make(map[string]int)
	for i, dv := range dvs {
		key := string(rune('0'+dv.SectionID)) + "-" + dv.Property
		dvIndex[key] = i
	}
	
	for i := range result {
		for j, dv := range dvs {
			if result[i].ID == dv.SectionID {
				switch dv.Property {
				case "a":
					result[i].A = x[j]
				case "ix":
					result[i].Ix = x[j]
				case "iy":
					result[i].Iy = x[j]
				case "iz":
					result[i].Iz = x[j]
				case "e":
					result[i].E = x[j]
				case "rho":
					result[i].Rho = x[j]
				}
			}
		}
	}
	return result
}

func calculateMass(x []float64, elements []model.Element, sections []model.Section, dvs []model.DesignVariable) float64 {
	tempSections := updateSections(sections, dvs, x)
	return CalculateTotalMass(elements, tempSections)
}

func computeGradient(x []float64, request model.OptimizationRequest, dvs []model.DesignVariable) []float64 {
	grad := make([]float64, len(dvs))
	h := 1e-6
	
	for i := range dvs {
		xPlus := append([]float64(nil), x...)
		xPlus[i] += h
		massPlus := calculateMass(xPlus, request.Elements, request.Sections, dvs)
		
		xMinus := append([]float64(nil), x...)
		xMinus[i] -= h
		massMinus := calculateMass(xMinus, request.Elements, request.Sections, dvs)
		
		grad[i] = (massPlus - massMinus) / (2 * h)
	}
	return grad
}

func computeConstraintGradients(x []float64, request model.OptimizationRequest, dvs []model.DesignVariable) [][]float64 {
	h := 1e-6
	numConstraints := len(request.FrequencyConstraints)
	grads := make([][]float64, numConstraints)
	
	for i, fc := range request.FrequencyConstraints {
		grads[i] = make([]float64, len(dvs))
		
		for j := range dvs {
			xPlus := append([]float64(nil), x...)
			xPlus[j] += h
			tempSections := updateSections(request.Sections, dvs, xPlus)
			freqsPlus, _ := computeFrequencies(request.Nodes, request.Elements, tempSections, request.Constraints, request.NumModes)
			
			xMinus := append([]float64(nil), x...)
			xMinus[j] -= h
			tempSections = updateSections(request.Sections, dvs, xMinus)
			freqsMinus, _ := computeFrequencies(request.Nodes, request.Elements, tempSections, request.Constraints, request.NumModes)
			
			if fc.ModeIndex-1 < len(freqsPlus) && fc.ModeIndex-1 < len(freqsMinus) {
				grads[i][j] = (freqsPlus[fc.ModeIndex-1] - freqsMinus[fc.ModeIndex-1]) / (2 * h)
			}
		}
	}
	return grads
}

func evaluateConstraints(freqs []float64, constraints []model.FrequencyConstraint) []float64 {
	violations := make([]float64, len(constraints))
	
	for i, fc := range constraints {
		if fc.ModeIndex-1 >= len(freqs) {
			continue
		}
		freq := freqs[fc.ModeIndex-1]
		
		switch fc.Type {
		case "lower":
			if freq < fc.LowerBound {
				violations[i] = fc.LowerBound - freq
			}
		case "upper":
			if freq > fc.UpperBound {
				violations[i] = freq - fc.UpperBound
			}
		case "both":
			if freq < fc.LowerBound {
				violations[i] = fc.LowerBound - freq
			} else if freq > fc.UpperBound {
				violations[i] = freq - fc.UpperBound
			}
		}
	}
	return violations
}

func updateDesignVariables(x, grad, constraintViolations []float64, lb, ub []float64, 
	constraints []model.FrequencyConstraint, freqs []float64) []float64 {
	newX := append([]float64(nil), x...)
	
	stepSize := 0.01
	
	for i := range newX {
		descent := grad[i]
		newX[i] -= stepSize * descent
	}
	
	for i, fc := range constraints {
		if fc.ModeIndex-1 >= len(freqs) {
			continue
		}
		freq := freqs[fc.ModeIndex-1]
		
		switch fc.Type {
		case "lower":
			if freq < fc.LowerBound {
				for j := range newX {
					sens := estimateSensitivity(j, fc.ModeIndex-1, x, constraints, freqs)
					if sens > 0 {
						violation := fc.LowerBound - freq
						newX[j] += violation / sens * 0.1
					}
				}
			}
		case "upper":
			if freq > fc.UpperBound {
				for j := range newX {
					sens := estimateSensitivity(j, fc.ModeIndex-1, x, constraints, freqs)
					if sens < 0 {
						violation := freq - fc.UpperBound
						newX[j] += violation / sens * 0.1
					}
				}
			}
		}
	}
	
	return newX
}

func estimateSensitivity(dvIndex, modeIndex int, x []float64, constraints []model.FrequencyConstraint, freqs []float64) float64 {
	h := 1e-6
	xPlus := append([]float64(nil), x...)
	xPlus[dvIndex] += h
	
	return 0.01
}

func PerformParamScan(request model.ParamScanRequest) (*model.ParamScanResponse, error) {
	if request.SecondVariable != nil {
		return perform2DParamScan(request)
	}
	
	step := (request.ScanEnd - request.ScanStart) / float64(request.NumSteps-1)
	scanValues := make([]float64, request.NumSteps)
	frequencies := make([][]float64, request.NumSteps)
	
	for i := 0; i < request.NumSteps; i++ {
		scanValues[i] = request.ScanStart + step*float64(i)
		
		tempSections := make([]model.Section, len(request.Sections))
		copy(tempSections, request.Sections)
		
		for j := range tempSections {
			if tempSections[j].ID == request.DesignVariable.SectionID {
				switch request.DesignVariable.Property {
				case "a":
					tempSections[j].A = scanValues[i]
				case "ix":
					tempSections[j].Ix = scanValues[i]
				case "iy":
					tempSections[j].Iy = scanValues[i]
				case "iz":
					tempSections[j].Iz = scanValues[i]
				case "e":
					tempSections[j].E = scanValues[i]
				case "rho":
					tempSections[j].Rho = scanValues[i]
				}
			}
		}
		
		freqs, _ := computeFrequencies(request.Nodes, request.Elements, tempSections, request.Constraints, request.NumModes)
		frequencies[i] = freqs
	}
	
	return &model.ParamScanResponse{
		Success:     true,
		Message:     "参数扫描完成",
		ScanValues:  scanValues,
		Frequencies: frequencies,
	}, nil
}

func perform2DParamScan(request model.ParamScanRequest) (*model.ParamScanResponse, error) {
	if request.SecondVariable == nil || request.SecondScanStart == nil || 
		request.SecondScanEnd == nil || request.SecondNumSteps == nil {
		return nil, nil
	}
	
	step1 := (request.ScanEnd - request.ScanStart) / float64(request.NumSteps-1)
	step2 := (*request.SecondScanEnd - *request.SecondScanStart) / float64(*request.SecondNumSteps-1)
	
	scanValues1 := make([]float64, request.NumSteps)
	scanValues2 := make([]float64, *request.SecondNumSteps)
	
	for i := 0; i < request.NumSteps; i++ {
		scanValues1[i] = request.ScanStart + step1*float64(i)
	}
	for i := 0; i < *request.SecondNumSteps; i++ {
		scanValues2[i] = *request.SecondScanStart + step2*float64(i)
	}
	
	frequencies2D := make([][]float64, request.NumSteps)
	
	for i := 0; i < request.NumSteps; i++ {
		frequencies2D[i] = make([]float64, *request.SecondNumSteps)
		
		for j := 0; j < *request.SecondNumSteps; j++ {
			tempSections := make([]model.Section, len(request.Sections))
			copy(tempSections, request.Sections)
			
			for k := range tempSections {
				if tempSections[k].ID == request.DesignVariable.SectionID {
					switch request.DesignVariable.Property {
					case "a":
						tempSections[k].A = scanValues1[i]
					case "ix":
						tempSections[k].Ix = scanValues1[i]
					case "iy":
						tempSections[k].Iy = scanValues1[i]
					case "iz":
						tempSections[k].Iz = scanValues1[i]
					case "e":
						tempSections[k].E = scanValues1[i]
					case "rho":
						tempSections[k].Rho = scanValues1[i]
					}
				}
				if tempSections[k].ID == request.SecondVariable.SectionID {
					switch request.SecondVariable.Property {
					case "a":
						tempSections[k].A = scanValues2[j]
					case "ix":
						tempSections[k].Ix = scanValues2[j]
					case "iy":
						tempSections[k].Iy = scanValues2[j]
					case "iz":
						tempSections[k].Iz = scanValues2[j]
					case "e":
						tempSections[k].E = scanValues2[j]
					case "rho":
						tempSections[k].Rho = scanValues2[j]
					}
				}
			}
			
			freqs, _ := computeFrequencies(request.Nodes, request.Elements, tempSections, request.Constraints, request.NumModes)
			if len(freqs) > 0 {
				frequencies2D[i][j] = freqs[0]
			}
		}
	}
	
	return &model.ParamScanResponse{
		Success:           true,
		Message:           "二维参数扫描完成",
		ScanValues:        scanValues1,
		SecondScanValues:  scanValues2,
		Frequencies2D:     frequencies2D,
	}, nil
}