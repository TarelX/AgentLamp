import React from 'react';
import ReactDOM from 'react-dom/client';
import MainWindow from './MainWindow';
import Lamp from './Lamp';

const isLamp = window.location.pathname.startsWith('/lamp');

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>{isLamp ? <Lamp /> : <MainWindow />}</React.StrictMode>
);
