import React from 'react';
import './Button.css';

export interface ButtonProps {
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'icon' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  loading?: boolean;
  onClick?: () => void;
  type?: 'button' | 'submit' | 'reset';
  className?: string;
  ariaLabel?: string;
}

export function Button({
  children,
  variant = 'primary',
  size = 'md',
  disabled = false,
  loading = false,
  onClick,
  type = 'button',
  className = '',
  ariaLabel,
}: ButtonProps) {
  const baseClass = 'button';
  const variantClass = `button--${variant}`;
  const sizeClass = `button--${size}`;
  const disabledClass = disabled || loading ? 'button--disabled' : '';
  const loadingClass = loading ? 'button--loading' : '';

  const buttonClass = [
    baseClass,
    variantClass,
    sizeClass,
    disabledClass,
    loadingClass,
    className,
  ].filter(Boolean).join(' ');

  const handleClick = () => {
    if (!disabled && !loading && onClick) {
      onClick();
    }
  };

  return (
    <button
      type={type}
      className={buttonClass}
      onClick={handleClick}
      disabled={disabled || loading}
      aria-label={ariaLabel}
    >
      {loading && <span className="button__spinner" />}
      <span className={`button__content ${loading ? 'button__content--loading' : ''}`}>
        {children}
      </span>
    </button>
  );
}