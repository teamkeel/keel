import * as tracing from "./tracing";
import jwt from "jsonwebtoken";

/**
 * Builds the headers for the API request including authentication.
 * @param {Object|null} identity Optional identity object for authentication
 * @param {string|null} authToken Optional auth token for authentication
 * @returns {Object} Headers object
 */
function buildHeaders(identity, authToken) {
  const headers = { "Content-Type": "application/json" };

  // An Identity instance is provided - make a JWT
  if (identity !== null) {
    const base64pk = process.env.KEEL_PRIVATE_KEY;
    let privateKey = undefined;

    if (base64pk) {
      privateKey = Buffer.from(base64pk, "base64").toString("utf8");
    }

    headers["Authorization"] =
      "Bearer " +
      jwt.sign({}, privateKey, {
        algorithm: privateKey ? "RS256" : "none",
        expiresIn: 60 * 60 * 24,
        subject: identity.id,
        issuer: "https://keel.so",
      });
  }

  // If an auth token is provided that can be sent as-is
  if (authToken !== null) {
    headers["Authorization"] = "Bearer " + authToken;
  }

  return headers;
}

/**
 * Gets the API URL from environment variables.
 * @returns {string} The API URL
 * @throws {Error} If the API URL is not set
 */
function getApiUrl() {
  const apiUrl = process.env.KEEL_API_URL;
  if (!apiUrl) {
    throw new Error("KEEL_API_URL environment variable is not set");
  }
  return apiUrl;
}

/**
 * Task represents a task instance with action methods.
 */
class Task {
  /**
   * @param {Object} data The task data from the API
   * @param {string} taskName The name of the task/topic
   * @param {Object|null} identity Optional identity object for authentication
   * @param {string|null} authToken Optional auth token for authentication
   */
  constructor(data, taskName, identity = null, authToken = null) {
    this.id = data.id;
    this.topic = data.name;
    this.status = data.status;
    this.deferredUntil = data.deferredUntil
      ? new Date(data.deferredUntil)
      : undefined;
    this.createdAt = new Date(data.createdAt);
    this.updatedAt = new Date(data.updatedAt);
    this.assignedTo = data.assignedTo;
    this.assignedAt = data.assignedAt ? new Date(data.assignedAt) : undefined;
    this.resolvedAt = data.resolvedAt ? new Date(data.resolvedAt) : undefined;
    this.flowRunId = data.flowRunId;

    // Store auth context for action methods
    this._taskName = taskName;
    this._identity = identity;
    this._authToken = authToken;
  }

  /**
   * Returns a new Task instance that will use the given identity for authentication.
   * @param {Object} identity The identity object
   * @returns {Task} A new Task instance with the identity set
   */
  withIdentity(identity) {
    const data = this._toApiData();
    return new Task(data, this._taskName, identity, null);
  }

  /**
   * Returns a new Task instance that will use the given auth token for authentication.
   * @param {string} token The auth token to use
   * @returns {Task} A new Task instance with the auth token set
   */
  withAuthToken(token) {
    const data = this._toApiData();
    return new Task(data, this._taskName, null, token);
  }

  /**
   * Converts the task back to API data format for creating new instances.
   * @returns {Object} The task data in API format
   */
  _toApiData() {
    return {
      id: this.id,
      name: this.topic,
      status: this.status,
      deferredUntil: this.deferredUntil?.toISOString(),
      createdAt: this.createdAt.toISOString(),
      updatedAt: this.updatedAt.toISOString(),
      assignedTo: this.assignedTo,
      assignedAt: this.assignedAt?.toISOString(),
      resolvedAt: this.resolvedAt?.toISOString(),
      flowRunId: this.flowRunId,
    };
  }

  /**
   * Assigns the task to an identity.
   * @param {Object} options Options containing identityId
   * @param {string} options.identityId The ID of the identity to assign the task to
   * @returns {Promise<Task>} The updated task
   */
  async assign({ identityId }) {
    const name = tracing.spanNameForModelAPI(this._taskName, "assign");

    return tracing.withSpan(name, async () => {
      const apiUrl = getApiUrl();
      const url = `${apiUrl}/topics/json/${this._taskName}/tasks/${this.id}/assign`;

      const response = await fetch(url, {
        method: "PUT",
        headers: buildHeaders(this._identity, this._authToken),
        body: JSON.stringify({ assigned_to: identityId }),
      });

      if (!response.ok) {
        const errorBody = await response.json().catch(() => ({}));
        throw new Error(
          `Failed to assign task: ${response.status} ${response.statusText} - ${
            errorBody.message || JSON.stringify(errorBody)
          }`
        );
      }

      const result = await response.json();
      return new Task(result, this._taskName, this._identity, this._authToken);
    });
  }

  /**
   * Starts the task, creating and running the associated flow.
   * @returns {Promise<Task>} The updated task with flowRunId
   */
  async start() {
    const name = tracing.spanNameForModelAPI(this._taskName, "start");

    return tracing.withSpan(name, async () => {
      const apiUrl = getApiUrl();
      const url = `${apiUrl}/topics/json/${this._taskName}/tasks/${this.id}/start`;

      const response = await fetch(url, {
        method: "PUT",
        headers: buildHeaders(this._identity, this._authToken),
      });

      if (!response.ok) {
        const errorBody = await response.json().catch(() => ({}));
        throw new Error(
          `Failed to start task: ${response.status} ${response.statusText} - ${
            errorBody.message || JSON.stringify(errorBody)
          }`
        );
      }

      const result = await response.json();
      return new Task(result, this._taskName, this._identity, this._authToken);
    });
  }

  /**
   * Completes the task.
   * @returns {Promise<Task>} The updated task
   */
  async complete() {
    const name = tracing.spanNameForModelAPI(this._taskName, "complete");

    return tracing.withSpan(name, async () => {
      const apiUrl = getApiUrl();
      const url = `${apiUrl}/topics/json/${this._taskName}/tasks/${this.id}/complete`;

      const response = await fetch(url, {
        method: "PUT",
        headers: buildHeaders(this._identity, this._authToken),
      });

      if (!response.ok) {
        const errorBody = await response.json().catch(() => ({}));
        throw new Error(
          `Failed to complete task: ${response.status} ${
            response.statusText
          } - ${errorBody.message || JSON.stringify(errorBody)}`
        );
      }

      const result = await response.json();
      return new Task(result, this._taskName, this._identity, this._authToken);
    });
  }

  /**
   * Defers the task until a specified date.
   * @param {Object} options Options containing deferUntil
   * @param {Date} options.deferUntil The date to defer the task until
   * @returns {Promise<Task>} The updated task
   */
  async defer({ deferUntil }) {
    const name = tracing.spanNameForModelAPI(this._taskName, "defer");

    return tracing.withSpan(name, async () => {
      const apiUrl = getApiUrl();
      const url = `${apiUrl}/topics/json/${this._taskName}/tasks/${this.id}/defer`;

      const response = await fetch(url, {
        method: "PUT",
        headers: buildHeaders(this._identity, this._authToken),
        body: JSON.stringify({ defer_until: deferUntil.toISOString() }),
      });

      if (!response.ok) {
        const errorBody = await response.json().catch(() => ({}));
        throw new Error(
          `Failed to defer task: ${response.status} ${response.statusText} - ${
            errorBody.message || JSON.stringify(errorBody)
          }`
        );
      }

      const result = await response.json();
      return new Task(result, this._taskName, this._identity, this._authToken);
    });
  }

  /**
   * Cancels the task.
   * @returns {Promise<Task>} The updated task
   */
  async cancel() {
    const name = tracing.spanNameForModelAPI(this._taskName, "cancel");

    return tracing.withSpan(name, async () => {
      const apiUrl = getApiUrl();
      const url = `${apiUrl}/topics/json/${this._taskName}/tasks/${this.id}/cancel`;

      const response = await fetch(url, {
        method: "PUT",
        headers: buildHeaders(this._identity, this._authToken),
      });

      if (!response.ok) {
        const errorBody = await response.json().catch(() => ({}));
        throw new Error(
          `Failed to cancel task: ${response.status} ${response.statusText} - ${
            errorBody.message || JSON.stringify(errorBody)
          }`
        );
      }

      const result = await response.json();
      return new Task(result, this._taskName, this._identity, this._authToken);
    });
  }
}

/**
 * TaskAPI provides methods for creating tasks via the HTTP API.
 */
class TaskAPI {
  /**
   * @param {string} taskName The name of the task/topic
   * @param {Object|null} identity Optional identity object for authentication
   * @param {string|null} authToken Optional auth token for authentication
   */
  constructor(taskName, identity = null, authToken = null) {
    this._taskName = taskName;
    this._identity = identity;
    this._authToken = authToken;
  }

  /**
   * Returns a new TaskAPI instance that will use the given identity for authentication.
   * @param {Object} identity The identity object
   * @returns {TaskAPI} A new TaskAPI instance with the identity set
   */
  withIdentity(identity) {
    return new TaskAPI(this._taskName, identity, null);
  }

  /**
   * Returns a new TaskAPI instance that will use the given auth token for authentication.
   * @param {string} token The auth token to use
   * @returns {TaskAPI} A new TaskAPI instance with the auth token set
   */
  withAuthToken(token) {
    return new TaskAPI(this._taskName, null, token);
  }

  /**
   * Creates a new task with the given data by calling the tasks API.
   * @param {Object} data The task data fields
   * @param {Object} options Optional settings like deferredUntil
   * @returns {Promise<Task>} The created task
   */
  async create(data = {}, options = {}) {
    const name = tracing.spanNameForModelAPI(this._taskName, "create");

    return tracing.withSpan(name, async () => {
      const apiUrl = getApiUrl();
      const url = `${apiUrl}/topics/json/${this._taskName}/tasks`;

      const body = {
        data: data,
      };

      if (options.deferredUntil) {
        body.defer_until = options.deferredUntil.toISOString();
      }

      const response = await fetch(url, {
        method: "POST",
        headers: buildHeaders(this._identity, this._authToken),
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorBody = await response.json().catch(() => ({}));
        throw new Error(
          `Failed to create task: ${response.status} ${response.statusText} - ${
            errorBody.message || JSON.stringify(errorBody)
          }`
        );
      }

      const result = await response.json();
      return new Task(result, this._taskName, this._identity, this._authToken);
    });
  }
}

export { TaskAPI, Task };
