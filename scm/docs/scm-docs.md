Zeus ERP: SCM Module Specification & Implementation Roadmap
1. Module Overview
The SCM module is the central router of the Zeus ERP
. It governs the acquisition of raw materials from vendors to fulfill manufacturing deficits and manages the outbound distribution of finished goods to clients
.
Core Objectives:
Procurement Intelligence: Automate vendor selection based on performance
.
Inventory Integrity: Prevent data discrepancies through "Blind Receiving" and immutable ledgers
.
Concurrency Control: Use "Eager Slot-Locking" to prevent duplicate procurement or shipping
.

--------------------------------------------------------------------------------
2. Core Subsystems & Logic
A. Vendor Routing (The Sourcing Brain)
Function: Maps MRP deficits to the best-fit supplier
.
Selection Logic: Uses a dynamic formula based on On-Time Rate and Defect Rate
.
Data Structure: Requires a Vendor-SKU Mapping table containing unit price, lead time, and minimum order quantities (MOQ)
.
B. Purchase Order (PO) Orchestration
The Mono-Vendor Rule: Each PO is a legal contract with a single vendor; mixed-vendor deficits must be split into separate drafts
.
Eager Slot-Locking: When an operator drafts a PO, the system "locks" that quantity from the global deficit pool for 30 minutes to prevent double-buying
.
State Machine: POs move linearly: Draft → Approved → In Transit → Received/Partial
. Regression (moving backward) is strictly prohibited
.
C. Goods Receipt (Inbound Logistics)
Blind Receiving: Operators must manually type counts; the system provides no "pre-filled" or "Receive All" buttons
.
Automated Quarantine: Components like batteries are flagged as defective if they exceed Aging Thresholds (e.g., 5-year-old stock) based on required production date entry
.
Atomic Updates: Upon completion, the system must simultaneously update the Inventory Ledger, the Inventory Snapshot, and the parent PO status
.
D. Downstream Logistics (Outbound)
Dispatch Orchestration: Ingests finalized orders from Sales and groups them by destination for carrier efficiency
.
Dispatch-Locking: Clicking "Pack Order" triggers a 30-minute lock to prevent duplicate shipping of the same order
.
Final Trigger: Transitioning to "Dispatched" status automatically triggers the Inventory Ledger to decrement the finished goods balance
.

--------------------------------------------------------------------------------
3. Implementation Roadmap
Phase 1: Foundation & Vendor Intelligence
Develop Supplier Registry: Implement CRUD for Supplier Profiles and Vendor-SKU Mappings
.
Performance Engine: Build the background service to calculate On-Time and Defect rates from historical receipt data
.
Phase 2: Inbound Procurement Engine
Deficit Aggregator: Create a service that listens to MRP "Shortage" alerts and pools them by SKU
.
PO Lifecycle Manager: Implement the one-way state machine and the 30-minute Eager Slot-Lock logic
.
Financial Handoff: Build the API to transmit total order values and payment terms to the Financial Module
.
Phase 3: Warehouse Ingestion (Receipts)
The "Blind" Interface: Build a UI that forces manual quantity and production date entry
.
Validation Logic: Implement the Aging Threshold check for sensitive components
.
The Atomic Ledger Bridge: Develop the transactional sequence to update Inventory Ledger (Inbound) and PO Status
.
Phase 4: Outbound Logistics
Sales Integration: Build the API endpoint to ingest "Finalized Sales Orders"
.
Packing Workflow: Implement the Dispatch-Locking procedure for warehouse operators
.
Carrier Integration: Create fields for Freight Data Capture (Carrier ID, Tracking #)
.
Inventory Sync: Ensure the "Dispatched" state triggers a ledger decrement
.

--------------------------------------------------------------------------------
4. Technical Constraints for AI Implementation
Concurrency: Use atomic session locks for all physical movement tasks (30 mins for packing/procurement, 60 mins for receiving)
.
Immutable Ledger: The SCM module writes to the ledger but the MRP module only reads it during readiness checks
.
Headless First: Outbound logistics must be able to receive orders via API from external B2B/B2C sources
.
5. Key Metrics (KPIs) to Track
Supplier Performance: On-Time Delivery % and Defect %
.
Warehouse Accuracy: Discrepancy counts between PO manifests and Blind Receipts
.
Fulfillment Speed: Time elapsed from "Pending" in Sales to "Dispatched" in SCM
.