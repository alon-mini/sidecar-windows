import React, {useEffect, useState, useRef} from 'react';

const THEMES = [
  {id: 'molokai', label: 'Molokai', color: '#a6e22e'},
  {id: 'nord', label: 'Nord', color: '#88C0D0'},
  {id: 'solarized-dark', label: 'Solarized', color: '#2AA198'},
  {id: 'tokyo-night', label: 'Tokyo', color: '#7AA2F7'},
];

export default function ThemeSwitcherNavbarItem() {
  const [theme, setTheme] = useState('molokai');
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef(null);

  useEffect(() => {
    const stored = localStorage.getItem('sidecar-theme') || 'molokai';
    setTheme(stored);
    document.documentElement.setAttribute('data-custom-theme', stored);

    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleThemeChange = (newTheme) => {
    setTheme(newTheme);
    localStorage.setItem('sidecar-theme', newTheme);
    document.documentElement.setAttribute('data-custom-theme', newTheme);
    setIsOpen(false);
  };

  const currentThemeObj = THEMES.find(t => t.id === theme) || THEMES[0];

  return (
    <div
      className="navbar__item dropdown"
      ref={dropdownRef}
    >
      <button
        className="navbar__link"
        onClick={(e) => {
          e.preventDefault();
          setIsOpen(!isOpen);
        }}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          fontFamily: 'var(--ifm-font-family-monospace)',
          fontSize: '13px',
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          padding: 'var(--ifm-navbar-item-padding-vertical) var(--ifm-navbar-item-padding-horizontal)',
          color: 'var(--ifm-navbar-link-color)',
          pointerEvents: 'auto',
        }}
      >
        <div
          style={{
            width: '10px',
            height: '10px',
            borderRadius: '50%',
            backgroundColor: currentThemeObj.color,
            boxShadow: `0 0 6px ${currentThemeObj.color}66`
          }}
        />
        <span>{currentThemeObj.label}</span>
      </button>
      
      <ul
        className={`dropdown__menu ${isOpen ? 'dropdown__menu--show' : ''}`}
        style={{
          position: 'absolute',
          top: 'calc(100% - 4px)',
          right: 0,
          left: 'auto',
          zIndex: 100,
          display: isOpen ? 'block' : 'none',
          opacity: isOpen ? 1 : 0,
          visibility: isOpen ? 'visible' : 'hidden',
          minWidth: '160px',
          padding: '12px 0 8px',
          margin: 0,
          listStyle: 'none',
          background: 'var(--sc-panel-2, #1b1f26)',
          border: '1px solid var(--sc-border, rgba(255,255,255,0.08))',
          borderRadius: '8px',
          boxShadow: '0 8px 24px rgba(0, 0, 0, 0.4)',
          pointerEvents: 'auto',
        }}
      >
        {THEMES.map((t) => (
          <li key={t.id}>
            <button
              className={`dropdown__link ${theme === t.id ? 'dropdown__link--active' : ''}`}
              onClick={(e) => {
                e.preventDefault();
                handleThemeChange(t.id);
              }}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '10px',
                fontFamily: 'var(--ifm-font-family-monospace)',
                fontSize: '12px',
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                width: '100%',
                textAlign: 'left',
                pointerEvents: 'auto',
              }}
            >
              <div
                style={{
                  width: '8px',
                  height: '8px',
                  borderRadius: '50%',
                  backgroundColor: t.color,
                }}
              />
              {t.label}
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
