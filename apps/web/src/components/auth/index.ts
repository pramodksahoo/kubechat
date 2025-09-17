// Authentication components exports
// Following coding standards from docs/architecture/coding-standards.md

export { LoginForm } from './LoginForm';
export { RegisterForm } from './RegisterForm';
export { ProtectedRoute, withAuth, useRequireAuth } from './ProtectedRoute';

export type {
  LoginFormProps,
} from './LoginForm';

export type {
  RegisterFormProps,
} from './RegisterForm';

export type {
  ProtectedRouteProps,
} from './ProtectedRoute';