# SyncWrite - Real-Time Collaborative Document Editor

SyncWrite is a web-based document editor that allows multiple users to collaborate on rich text documents in real-time. The application ensures low-latency synchronization of document changes, efficient conflict resolution, and a seamless user experience. It leverages modern technologies to provide a scalable, secure, and responsive platform for collaborative editing.

> ![NOTE]
> This is just outline of project plan. More data to be added sooner. 

## Tech Stack

- **Backend**:
  - **Go (Golang)**: Backend programming language for high performance and concurrency.
  - **gRPC**: High-performance RPC framework for communication between backend services.
  - **MQTT**: Lightweight messaging protocol for real-time communication and synchronization.
  - **Protocol Buffers**: Serialization format used by gRPC for defining service contracts.
  - **Redis**: In-memory data store for session management and real-time data.
  - **PostgreSQL**: Relational database for persistent storage of user data and documents.

- **Frontend**:
  - **SolidJS**: A modern, efficient, and reactive JavaScript library for building user interfaces.
  - **Bulma CSS**: Modern CSS framework based on Flexbox for rapid UI development.
  - **TypeScript**: Superset of JavaScript for type safety and better tooling.
  - **MQTT.js**: MQTT client library for JavaScript to enable MQTT over WebSockets.
  - **gRPC-Web**: Allows web applications to communicate with gRPC backend services.

- **Other Tools**:
  - **Docker**: Containerization for consistent development and deployment environments.
  - **Nginx**: Web server and reverse proxy for serving the application.
  - **Let's Encrypt**: For securing communication with SSL/TLS certificates.

## Features

- **Real-Time Collaboration**: Instantaneous synchronization of document edits among all collaborators.
- **Rich Text Editing**: Support for text formatting, images, lists, tables, and more.
- **User Authentication**: Secure registration and login with session management.
- **Access Control**: Document sharing with customizable permissions (owner, editor, viewer).
- **Version History**: Ability to track changes and revert to previous document versions.
- **User Presence Indicators**: Visual cues showing who is currently editing the document.
- **Conflict Resolution**: Efficient handling of concurrent edits using CRDTs (Conflict-Free Replicated Data Types).
- **Scalability**: Designed to handle a large number of concurrent users and documents.
