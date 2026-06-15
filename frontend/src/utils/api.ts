import axios from 'axios'
import type { AnalysisRequest, AnalysisResponse, Preset, FRFRequest, FRFResponse, TransientRequest, TransientResponse } from '../types'

const BASE_URL = 'http://localhost:8080/api'

export const api = {
    analyze: async (request: AnalysisRequest): Promise<AnalysisResponse> => {
        const response = await axios.post(`${BASE_URL}/analyze`, request)
        return response.data
    },

    getPresets: async (): Promise<Preset[]> => {
        const response = await axios.get(`${BASE_URL}/presets`)
        return response.data
    },

    loadPreset: async (name: string): Promise<any> => {
        const response = await axios.post(`${BASE_URL}/preset/${name}`)
        return response.data
    },

    calculateFRF: async (request: FRFRequest): Promise<FRFResponse> => {
        const response = await axios.post(`${BASE_URL}/frf`, request)
        return response.data
    },

    calculateTransient: async (request: TransientRequest): Promise<TransientResponse> => {
        const response = await axios.post(`${BASE_URL}/transient`, request)
        return response.data
    }
}
