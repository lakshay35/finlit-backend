package models

type FitnessHistory struct {
	TotalRecords int                    `json:"total_records,omitempty"`
	TotalPages   int                    `json:"total_pages,omitempty"`
	PageIndex    int                    `json:"page_index,omitempty"`
	Records      []FitnessHistoryRecord `json:"records"`
	Month        int                    `json:"month,omitempty"`
}
