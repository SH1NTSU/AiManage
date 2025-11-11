export interface PasswordValidation {
  isValid: boolean;
  errors: string[];
  requirements: {
    minLength: boolean;
    hasNumber: boolean;
    hasLetter: boolean;
  };
}

export const validatePassword = (password: string): PasswordValidation => {
  const requirements = {
    minLength: password.length >= 8,
    hasNumber: /\d/.test(password),
    hasLetter: /[a-zA-Z]/.test(password),
  };

  const errors: string[] = [];

  if (!requirements.minLength) {
    errors.push("Password must be at least 8 characters long");
  }
  if (!requirements.hasNumber) {
    errors.push("Password must contain at least one number");
  }
  if (!requirements.hasLetter) {
    errors.push("Password must contain at least one letter");
  }

  return {
    isValid: requirements.minLength && requirements.hasNumber && requirements.hasLetter,
    errors,
    requirements,
  };
};
