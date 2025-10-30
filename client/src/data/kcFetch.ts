interface RawRequestError {
  message: string;
  code?: number;
  details?: string;
}

class ApiRequestError extends Error implements RawRequestError {
  message: string;

  code?: number;

  details?: string;

  constructor(jsonResponse: RawRequestError = {} as RawRequestError) {
    const {
      message = '',
      code = 0,
      details = '',
    } = jsonResponse;

    super();
    this.message = message;
    this.code = code;
    this.details = details;
  }
}

 
const kcFetch = (url: string, options?: RequestInit) => fetch(url, {...options })
  .then(async (response: Response) => {
    const contentType = response.headers?.get('Content-Type');
    if (!response.ok) {
      if (contentType && contentType.includes('application/json')) {
        // handle JSON error response
        const errorResult = await response.json();
        if (!errorResult.code) {
          errorResult.code = response.status;
        }

        throw new ApiRequestError(errorResult);
      }
      throw new ApiRequestError();
    }
    if(contentType && contentType.includes('text/plain')) {
      return (await response.blob()).text();
    }
    if(contentType && contentType.includes('application/json')) {
      return response.json();
    }
    return;
  });

export { ApiRequestError };
export type { RawRequestError };
export default kcFetch;
