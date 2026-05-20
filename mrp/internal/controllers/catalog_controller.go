package controllers

import "net/http"

// GET /api/v1/mrp/assemblies
func (c *ProductionController) GetAssemblies(w http.ResponseWriter, r *http.Request) {
	// TODO: Call c.svc.GetAssemblies
}

// POST /api/v1/mrp/assemblies
func (c *ProductionController) CreateAssembly(w http.ResponseWriter, r *http.Request) {
	// TODO: Decode models.CreateAssemblyRequest and call c.svc.CreateAssembly
}

// PUT /api/v1/mrp/assemblies/{id}
func (c *ProductionController) UpdateAssembly(w http.ResponseWriter, r *http.Request) {
	// TODO: Decode models.UpdateAssemblyRequest and call c.svc.UpdateAssembly
}

// DELETE /api/v1/mrp/assemblies/{id}
func (c *ProductionController) DeleteAssembly(w http.ResponseWriter, r *http.Request) {
	// TODO: Call c.svc.DeleteAssembly
}

// GET /api/v1/mrp/catalog
func (c *ProductionController) GetCatalog(w http.ResponseWriter, r *http.Request) {
	// TODO: Call c.svc.GetCatalog
}

// GET /api/v1/mrp/catalog/{sku}/where-used
func (c *ProductionController) GetWhereUsed(w http.ResponseWriter, r *http.Request) {
	// TODO: Call c.svc.GetWhereUsed
}
