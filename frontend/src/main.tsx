import React from 'react';
import ReactDOM from 'react-dom/client';
import Lamp from './Lamp';
import Settings from './Settings';

const isSettings = window.location.pathname.startsWith('/settings');

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>{isSettings ? <Settings /> : <Lamp />}</React.StrictMode>
);
