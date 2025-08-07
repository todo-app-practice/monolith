export interface User {
  Id: number;
  CreatedAt: string;
  UpdatedAt: string;
  FirstName: string;
  LastName: string;
  Email: string;
  IsEmailVerified: boolean;
}

export interface Todo {
  Id: number;
  CreatedAt: string;
  UpdatedAt: string;
  Text: string;
  Done: boolean;
  UserId: number;
} 