package solver

import (
    "math"
    "modal-analysis/internal/model"
)

type Matrix struct {
    rows, cols int
    data       []float64
}

func NewMatrix(rows, cols int) *Matrix {
    return &Matrix{
        rows: rows,
        cols: cols,
        data: make([]float64, rows*cols),
    }
}

func (m *Matrix) At(i, j int) float64 {
    return m.data[i*m.cols+j]
}

func (m *Matrix) Set(i, j int, val float64) {
    m.data[i*m.cols+j] = val
}

func (m *Matrix) Rows() int {
    return m.rows
}

func (m *Matrix) Cols() int {
    return m.cols
}

func (m *Matrix) Add(other *Matrix) {
    for i := 0; i < m.rows*m.cols; i++ {
        m.data[i] += other.data[i]
    }
}

func (m *Matrix) Scale(factor float64) {
    for i := range m.data {
        m.data[i] *= factor
    }
}

func (m *Matrix) Multiply(other *Matrix) *Matrix {
    result := NewMatrix(m.rows, other.cols)
    for i := 0; i < m.rows; i++ {
        for j := 0; j < other.cols; j++ {
            var sum float64
            for k := 0; k < m.cols; k++ {
                sum += m.At(i, k) * other.At(k, j)
            }
            result.Set(i, j, sum)
        }
    }
    return result
}

func (m *Matrix) Transpose() *Matrix {
    result := NewMatrix(m.cols, m.rows)
    for i := 0; i < m.rows; i++ {
        for j := 0; j < m.cols; j++ {
            result.Set(j, i, m.At(i, j))
        }
    }
    return result
}

func (m *Matrix) Copy() *Matrix {
    result := NewMatrix(m.rows, m.cols)
    copy(result.data, m.data)
    return result
}

func (m *Matrix) Diagonal() []float64 {
    result := make([]float64, min(m.rows, m.cols))
    for i := 0; i < len(result); i++ {
        result[i] = m.At(i, i)
    }
    return result
}

func (m *Matrix) ApplyBC(fixedDofs []int) {
    for _, dof := range fixedDofs {
        for j := 0; j < m.cols; j++ {
            m.Set(dof, j, 0)
            m.Set(j, dof, 0)
        }
        m.Set(dof, dof, 1)
    }
}

func (m *Matrix) IsPositiveDefinite() bool {
    n := m.rows
    for i := 0; i < n; i++ {
        if m.At(i, i) <= 0 {
            return false
        }
    }
    return true
}

func (m *Matrix) VectorMultiply(v []float64) []float64 {
    result := make([]float64, m.rows)
    for i := 0; i < m.rows; i++ {
        var sum float64
        for j := 0; j < m.cols; j++ {
            sum += m.At(i, j) * v[j]
        }
        result[i] = sum
    }
    return result
}

func ComputeLocalStiffnessMatrix(section model.Section, length float64) *Matrix {
    k := NewMatrix(12, 12)
    E := section.E
    A := section.A
    Ix := section.Ix
    Iy := section.Iy
    
    EA_L := E * A / length
    EI_L3 := E * Ix / (length * length * length)
    EI_L2 := E * Ix / (length * length)
    EI_L := E * Ix / length
    
    GJ_L := E * Iy / (2 * (1 + section.Nu) * length)
    
    k.Set(0, 0, EA_L)
    k.Set(0, 6, -EA_L)
    k.Set(6, 0, -EA_L)
    k.Set(6, 6, EA_L)
    
    k.Set(1, 1, 12*EI_L3)
    k.Set(1, 2, 6*EI_L2)
    k.Set(1, 4, -6*EI_L2)
    k.Set(1, 7, -12*EI_L3)
    k.Set(1, 8, 6*EI_L2)
    k.Set(1, 10, -6*EI_L2)
    
    k.Set(2, 1, 6*EI_L2)
    k.Set(2, 2, 4*EI_L)
    k.Set(2, 4, 2*EI_L)
    k.Set(2, 7, -6*EI_L2)
    k.Set(2, 8, 2*EI_L)
    k.Set(2, 10, 4*EI_L)
    
    k.Set(3, 3, GJ_L)
    k.Set(3, 9, -GJ_L)
    k.Set(9, 3, -GJ_L)
    k.Set(9, 9, GJ_L)
    
    k.Set(4, 1, -6*EI_L2)
    k.Set(4, 2, 2*EI_L)
    k.Set(4, 4, 4*EI_L)
    k.Set(4, 7, 6*EI_L2)
    k.Set(4, 8, -2*EI_L)
    k.Set(4, 10, 2*EI_L)
    
    k.Set(5, 5, E*Iy/length)
    k.Set(5, 11, -E*Iy/length)
    k.Set(11, 5, -E*Iy/length)
    k.Set(11, 11, E*Iy/length)
    
    k.Set(7, 1, -12*EI_L3)
    k.Set(7, 2, -6*EI_L2)
    k.Set(7, 4, 6*EI_L2)
    k.Set(7, 7, 12*EI_L3)
    k.Set(7, 8, -6*EI_L2)
    k.Set(7, 10, 6*EI_L2)
    
    k.Set(8, 1, 6*EI_L2)
    k.Set(8, 2, 2*EI_L)
    k.Set(8, 4, -2*EI_L)
    k.Set(8, 7, -6*EI_L2)
    k.Set(8, 8, 4*EI_L)
    k.Set(8, 10, -2*EI_L)
    
    k.Set(10, 1, -6*EI_L2)
    k.Set(10, 2, 4*EI_L)
    k.Set(10, 4, 2*EI_L)
    k.Set(10, 7, 6*EI_L2)
    k.Set(10, 8, -2*EI_L)
    k.Set(10, 10, 4*EI_L)
    
    return k
}

func ComputeLocalMassMatrix(section model.Section, length float64) *Matrix {
    m := NewMatrix(12, 12)
    rho := section.Rho
    A := section.A
    Ix := section.Ix
    Iy := section.Iy
    mass := rho * A * length
    
    m.Set(0, 0, mass/2)
    m.Set(6, 6, mass/2)
    
    m.Set(1, 1, mass/2)
    m.Set(7, 7, mass/2)
    
    m.Set(2, 2, mass/2)
    m.Set(8, 8, mass/2)
    
    m.Set(3, 3, rho*Ix)
    m.Set(9, 9, rho*Ix)
    
    m.Set(4, 4, rho*Iy)
    m.Set(10, 10, rho*Iy)
    
    m.Set(5, 5, rho*Iy)
    m.Set(11, 11, rho*Iy)
    
    m.Set(1, 7, mass/2)
    m.Set(7, 1, mass/2)
    
    m.Set(2, 8, mass/2)
    m.Set(8, 2, mass/2)
    
    return m
}

func ComputeTransformationMatrix(node1, node2 model.Node) *Matrix {
    T := NewMatrix(12, 12)
    
    dx := node2.X - node1.X
    dy := node2.Y - node1.Y
    dz := node2.Z - node1.Z
    length := math.Sqrt(dx*dx + dy*dy + dz*dz)
    
    if length == 0 {
        for i := 0; i < 12; i++ {
            T.Set(i, i, 1)
        }
        return T
    }
    
    nx := dx / length
    ny := dy / length
    nz := dz / length
    
    var sx, sy, sz float64
    if math.Abs(nx) > math.Abs(ny) && math.Abs(nx) > math.Abs(nz) {
        sx = 0
        sy = nz
        sz = -ny
    } else {
        sx = nz
        sy = 0
        sz = -nx
    }
    
    slen := math.Sqrt(sx*sx + sy*sy + sz*sz)
    sx /= slen
    sy /= slen
    sz /= slen
    
    tx := ny*sz - nz*sy
    ty := nz*sx - nx*sz
    tz := nx*sy - ny*sx
    
    for i := 0; i < 4; i++ {
        T.Set(3*i, 3*i, nx)
        T.Set(3*i, 3*i+1, ny)
        T.Set(3*i, 3*i+2, nz)
        T.Set(3*i+1, 3*i, sx)
        T.Set(3*i+1, 3*i+1, sy)
        T.Set(3*i+1, 3*i+2, sz)
        T.Set(3*i+2, 3*i, tx)
        T.Set(3*i+2, 3*i+1, ty)
        T.Set(3*i+2, 3*i+2, tz)
    }
    
    return T
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
