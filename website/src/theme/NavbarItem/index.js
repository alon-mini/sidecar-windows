import React from 'react';
import NavbarItem from '@theme-original/NavbarItem';
import ThemeSwitcherNavbarItem from '@site/src/components/ThemeSwitcherNavbarItem';

export default function NavbarItemWrapper(props) {
  if (props.type === 'custom-themeSwitcher') {
    return <ThemeSwitcherNavbarItem {...props} />;
  }
  return <NavbarItem {...props} />;
}
