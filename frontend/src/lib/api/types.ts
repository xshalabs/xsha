export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  message: string;
  user: string;
  token: string;
}

export interface UserResponse {
  user: string;
  authenticated: boolean;
  message: string;
}

export interface ApiErrorResponse {
  error: string;
  details?: string;
}

export interface LanguagesResponse {
  message: string;
  languages: string[];
  current: string;
}
