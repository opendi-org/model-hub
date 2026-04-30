const API_URL = typeof __API_URL__ !== 'undefined' && __API_URL__
  ? __API_URL__
  : '/api';

export default API_URL;