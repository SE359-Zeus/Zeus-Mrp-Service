--------------------------------------------------------------------------------
Zeus ERP: System Service Module Implementation Specification
1. Module Overview
The System Service is the "Root of Trust" for the Zeus ERP. It is strictly restricted to Administrators and manages the authentication lifecycle, user directory, and a system-wide immutable audit trail
.
Core Objectives:
Centralized Identity: Single point for user authentication and session management
.
Decentralized Authorization: Use cryptographically signed tokens so downstream modules (Sales, MRP, SCM) can verify permissions without re-querying the database
.
Compliance Traceability: Provide a read-only ledger of every critical state change across the ecosystem
.

--------------------------------------------------------------------------------
2. Core Subsystems & Technical Logic
A. Authentication & Identity (The "Gatekeeper")
Mechanism: Uses an Asymmetric Algorithm (RS256 or EdDSA) to sign tokens
.
Token Strategy:
Access Token (Short-lived): A JWT containing User ID and Roles/Permissions
.
Refresh Token (Long-lived): Used to rotate access tokens without requiring manual login
.
Verification: The System Service holds the Private Key (to sign); all other modules (SCM, MRP, Sales) hold the Public Key to independently verify the signature
.
B. User Management (The Directory)
User Roles: Implements three functional roles: Admin, Editor, and Viewer
.
Account Lifecycle: Administrators can provision accounts and instantly revoke access by toggling a status to INACTIVE, which kills the ability to generate new tokens
.
Data Structure: User Profile tracks Email, Role, Last Login, and Account Status
.
C. Audit Logging (The Ledger)
Event Ingestion: Subsystems (Sales, MRP, SCM) emit asynchronous payloads to this module whenever a state change occurs
.
Asynchronicity: This ensures business transactions (like shipping an order) are not slowed down by the logging process
.
Event Record Structure: Must capture: Timestamp, User, Action Type, Target Resource, Details, and IP Address
.

--------------------------------------------------------------------------------
3. API & Integration Blueprint
I. Identity Endpoints
Endpoint
Method
Input
Output
Purpose
/auth/login
POST
Credentials
Access + Refresh Token
Authenticate user and issue signed JWT
.
/auth/refresh
POST
Refresh Token
New Access Token
Rotate session tokens
.
/users
POST/PUT
Profile Data
Success/Fail
Create or modify operator accounts
.
II. Audit Stream (Internal)
Payload: POST /logs/ingest (Internal API)
Logic: Accepts event data from other modules. Must be optimized for high-volume writes
.

--------------------------------------------------------------------------------
4. Implementation Roadmap for AI
Phase 1: Security Foundation
Key Pair Generation: Set up the asymmetric encryption (Private/Public keys)
.
JWT Middleware: Create the verification logic that the Sales, MRP, and SCM modules will use to validate tokens via the Public Key
.
Phase 2: User Registry
Database Schema: Implement the User Profile and Role structures
.
Admin UI: Build the User Access dashboard with the "Add New User" and "ACTIVE/INACTIVE" toggle functionality
.
Phase 3: Logging Infrastructure
Async Log Engine: Build the receiver for incoming event streams
.
Audit UI: Create the Audit Logs view with filters for Action Types (Login, Create, Update, Delete) and a "Live" stream toggle
.

--------------------------------------------------------------------------------
5. Critical Business Rules & Constraints
State Machine: Account Status is binary (ACTIVE vs. INACTIVE). An INACTIVE account must be blocked from all authentication attempts immediately
.
Read-Only Logs: The Audit Log must be immutable. No "Edit" or "Delete" functions should exist for event records
.
Restricted Access: Access to the System Service UI is limited exclusively to the Administrator role
.
Security Alerting: Any event tagged as a "Security Event" (e.g., failed logins, unauthorized resource access) must be highlighted in Red in the Audit Dashboard
.
6. Key Performance Indicators (KPIs)
Logins Today: Monitoring daily system activity
.
Security Events: Real-time count of failed or flagged operations
.
Modification Velocity: Tracking the total number of UPDATE and DELETE actions across the ERP
.