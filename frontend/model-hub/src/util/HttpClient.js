function handleResponse(response) {
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export default class HTTPClient {
  // React exposes env via process.env.REACT_APP_API_URL. When empty, use '/api' for proxy/Nginx.
  static baseURL = process.env.REACT_APP_API_URL || '/api';

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