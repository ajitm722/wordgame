import axios, { AxiosError, AxiosRequestConfig } from "axios";

/*
 * Single HTTP choke point — all API calls go through this function.
 * Accepts a type parameter T for the response shape, so callers
 * get typed responses without manual casting.
 */
export async function sendRequest<T>(
  method: "GET" | "POST",
  path: string,
  data?: unknown
): Promise<T> {
  const config: AxiosRequestConfig = {
    method,
    url: path,
    data,
    headers: { "Content-Type": "application/json" },
  };

  try {
    const response = await axios(config);
    return response.data as T;
  } catch (err) {
    const axiosErr = err as AxiosError<{ message?: string }>;
    throw new Error(
      axiosErr.response?.data?.message ||
        axiosErr.message ||
        "Network error"
    );
  }
}
