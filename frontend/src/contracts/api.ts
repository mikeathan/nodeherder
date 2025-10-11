import { ApiResponse } from "@/types/api.type";


export function isApiResponse<T>(res: any): res is ApiResponse<T> {
  return typeof res === 'object' && res !== null;
}
