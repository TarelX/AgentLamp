import React from 'react';
import ReactDOM from 'react-dom/client';
import MainWindow from './MainWindow';
import Lamp from './Lamp';

const params = new URLSearchParams(window.location.search);
const isLamp = params.get('mode') === 'lamp';

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>{isLamp ? <Lamp /> : <MainWindow />}</React.StrictMode>
);
