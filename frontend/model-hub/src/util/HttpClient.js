async function handleResponse(response) {
  const contentType = response.headers.get('content-type') || '';
  const isJson = contentType.includes('application/json');
  
  if (!response.ok) {
    if (isJson) {
      const errorData = await response.json().catch(() => ({}));
      if (errorData && errorData.error) {
        const err = new Error(errorData.error);
        if (Array.isArray(errorData.details)) {
          err.details = errorData.details;
        }
        throw err;
      }
      throw new Error(`HTTP ${response.status}`);
    }
    const text = await response.text();
    throw new Error(text || `HTTP ${response.status}`);
  }

  if (isJson) {
    return response.json();
  }
  const text = await response.text();
  return text ? { message: text } : {};
}

export default class HTTPClient {
  // Injected by Vite in build/dev. Falls back to relative /api behind reverse proxy.
  static baseURL = (typeof __API_URL__ !== 'undefined' && __API_URL__)
    ? __API_URL__
    : '/api';

  // GET request (credentials: include so session cookies are sent)
  static async get(url) {
    return fetch(this.baseURL + url, { credentials: 'include' }).then(handleResponse);
  }

  // POST request (credentials: include so session cookies are sent)
  static async post(url, data) {
    const response = await fetch(HTTPClient.baseURL + url, {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });
    return handleResponse(response);
  }

  // POST with FormData (e.g. file upload). Do not set Content-Type; browser sets multipart boundary.
  static async postFormData(url, formData) {
    const response = await fetch(HTTPClient.baseURL + url, {
      method: 'POST',
      credentials: 'include',
      body: formData,
    });
    return handleResponse(response);
  }

  // PUT request (credentials: include so session cookies are sent)
  static async put(url, data) {
    const response = await fetch(HTTPClient.baseURL + url, {
      method: 'PUT',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });
    return handleResponse(response);
  }

  // PATCH request (credentials: include so session cookies are sent)
  static async patch(url, data) {
    const response = await fetch(HTTPClient.baseURL + url, {
      method: 'PATCH',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });
    return handleResponse(response);
  }

  // DELETE request (credentials: include so session cookies are sent)
  static async delete(url) {
    const response = await fetch(HTTPClient.baseURL + url, {
      method: 'DELETE',
      credentials: 'include',
    });
    return handleResponse(response);
  }
}