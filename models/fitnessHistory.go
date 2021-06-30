package models

type FitnessHistory struct {
	TotalRecords int                    `json:"total_records"`
	TotalPages   int                    `json:"total_pages"`
	PageIndex    int                    `json:"page_index"`
	Records      []FitnessHistoryRecord `json:"records"`
}
