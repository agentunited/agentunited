import React, { forwardRef, useState } from 'react';
import './Input.css';

export interface InputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label?: string;
  error?: string;
  variant?: 'default' | 'search';
  size?: 'sm' | 'md' | 'lg';
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  onRightIconClick?: () => void;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(({
  label,
  error,
  variant = 'default',
  size = 'md',
  leftIcon,
  rightIcon,
  onRightIconClick,
  className = '',
  disabled = false,
  type = 'text',
  id,
  ...props
}, ref) => {
  const [isFocused, setIsFocused] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const inputId = id || label?.toLowerCase().replace(/\s+/g, '-');
  const isPassword = type === 'password';
  const actualType = isPassword && showPassword ? 'text' : type;

  const containerClass = [
    'input-container',
    `input-container--${variant}`,
    `input-container--${size}`,
    isFocused ? 'input-container--focused' : '',
    error ? 'input-container--error' : '',
    disabled ? 'input-container--disabled' : '',
    className,
  ].filter(Boolean).join(' ');

  const handleFocus = (e: React.FocusEvent<HTMLInputElement>) => {
    setIsFocused(true);
    props.onFocus?.(e);
  };

  const handleBlur = (e: React.FocusEvent<HTMLInputElement>) => {
    setIsFocused(false);
    props.onBlur?.(e);
  };

  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword);
  };

  const handleRightIconClick = () => {
    if (isPassword) {
      togglePasswordVisibility();
    } else if (onRightIconClick) {
      onRightIconClick();
    }
  };

  const rightIconElement = isPassword ? (
    <span className="input__icon input__icon--password">
      {showPassword ? '🙈' : '👁️'}
    </span>
  ) : rightIcon ? (
    <span className="input__icon">{rightIcon}</span>
  ) : null;

  return (
    <div className="input-wrapper">
      {label && (
        <label htmlFor={inputId} className="input__label">
          {label}
        </label>
      )}
      
      <div className={containerClass}>
        {leftIcon && (
          <span className="input__icon input__icon--left">
            {leftIcon}
          </span>
        )}
        
        <input
          ref={ref}
          id={inputId}
          type={actualType}
          className="input"
          disabled={disabled}
          onFocus={handleFocus}
          onBlur={handleBlur}
          aria-invalid={error ? 'true' : 'false'}
          aria-describedby={error ? `${inputId}-error` : undefined}
          {...props}
        />
        
        {rightIconElement && (
          <button
            type="button"
            className="input__icon-button"
            onClick={handleRightIconClick}
            disabled={disabled}
            aria-label={isPassword ? (showPassword ? 'Hide password' : 'Show password') : undefined}
          >
            {rightIconElement}
          </button>
        )}
      </div>
      
      {error && (
        <span id={`${inputId}-error`} className="input__error" role="alert">
          {error}
        </span>
      )}
    </div>
  );
});

Input.displayName = 'Input';