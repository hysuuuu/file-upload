# Robust Multipart: Distributed Large File Upload System

## Background

Most file upload projects stop at the "Happy Path" (Frontend Chunking -> Backend Receiving -> Assembling). However, in a real production environment, network jitter, client crashes, and concurrent write conflicts—the "Failure Modes" of distributed systems—are the true challenges.

This project aims to build a highly available, fault-tolerant, and cost-aware **production-grade large file upload system**. By implementing a rigorous state machine and reconciliation mechanism, it completely resolves data inconsistencies caused by relying on client-side state and defends against extreme edge cases of underlying storage APIs.

---

## Core Objectives

1. **Server-Side Source of Truth (Reconciliation):** Discard unreliable client-side `LocalStorage`. Let the backend and relational database drive the state verification for resumable uploads.
2. **Concurrency Control (Defense against Overwrites):** Resolve the silent data overwrite issue caused by S3's default `Last-Writer-Wins` behavior by implementing distributed locks and unique key strategies at the application layer.
3. **Lifecycle Management & Garbage Collection (GC):** Implement background scheduled tasks to clean up orphaned parts caused by network disconnections, preventing runaway cloud storage costs.
4. **Idempotency & Resilience:** Defend against the `HTTP 200-but-error` trap in distributed storage's `CompleteMultipartUpload` API, ensuring eventual data consistency.

---

## Tech Stack & Choices

| Module           | Technology                          | Why this?                                                                                                                                                                                |
| :--------------- | :---------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Backend**      | Node.js (TypeScript) / Go           | Needs to handle highly concurrent I/O-intensive requests. Node.js's async event loop or Go's goroutines provide excellent high-concurrency processing capabilities.                      |
| **Database**     | PostgreSQL                          | The system relies heavily on ACID properties to ensure dual-write consistency and relational queries between the "Upload Session Table" and the "Chunk State Table".                     |
| **Cache & Lock** | Redis                               | Acts as a cache to accelerate chunk state queries and implements Distributed Locks to resolve race conditions during concurrent multi-user uploads.                                      |
| **Storage**      | MinIO (Local) / AWS S3 (Production) | MinIO is fully compatible with the S3 API. It allows zero-cost local simulation of various failure modes in distributed object storage and seamless switching to AWS S3 upon deployment. |

---

## Core Workflows

### 1. State Reconciliation & Resumable Uploads

When a disconnected client reconnects, it does not rely on local memory.

- The client sends `GET /upload/{uploadId}/status`.
- The backend fetches the verified Part list from the DB (or directly queries S3 via `ListParts`).
- The backend returns the `successfully received list`, and the client only uploads the missing chunks after comparison.

### 2. Dual-Layer Data Integrity Validation

To prevent packet corruption during network transmission:

- The client attaches an `MD5/SHA256 Checksum` in the header when transmitting each Part.
- The storage layer (S3) automatically performs the first layer of validation; the backend records the Checksum in the DB and performs a second layer of logical verification before the final assembly.

### 3. Race Condition Defense

When two users upload updated versions of the same file simultaneously:

- The system requests a distributed lock for the `FileID` in Redis.
- Requests that fail to acquire the lock will immediately receive an `HTTP 409 Conflict`, prompting the user to retry or warning that the file is currently being edited.

### 4. Garbage Collection (GC)

To prevent incomplete uploads from occupying storage space indefinitely:

- Implement a Cron Job to periodically scan the DB for sessions with a `CreatedAt` older than 7 days (configurable TTL) and a `Pending` status.
- Proactively call S3's `AbortMultipartUpload` API to clean up the orphaned chunks and mark the DB state as `Failed/Aborted`.
