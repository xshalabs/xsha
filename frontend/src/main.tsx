import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'

// 异步初始化国际化
import { initializeI18n } from './i18n'

const rootElement = document.getElementById('root')!;
const root = createRoot(rootElement);

// 显示加载状态
const renderLoading = () => {
  root.render(
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      height: '100vh',
      fontSize: '16px',
      color: '#666'
    }}>
      Loading...
    </div>
  );
};

// 渲染应用
const renderApp = () => {
  root.render(
    <StrictMode>
      <App />
    </StrictMode>,
  );
};

// 初始化应用
const initApp = async () => {
  try {
    // 显示加载状态
    renderLoading();
    
    // 等待i18n初始化完成
    await initializeI18n();
    
    // 渲染应用
    renderApp();
  } catch (error) {
    console.error('Failed to initialize application:', error);
    
    // 渲染错误状态
    root.render(
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        fontSize: '16px',
        color: '#ff4444',
        textAlign: 'center',
        padding: '20px'
      }}>
        <div>
          <div>Failed to load application</div>
          <div style={{ fontSize: '14px', marginTop: '8px', opacity: 0.7 }}>
            Please refresh the page to try again
          </div>
        </div>
      </div>
    );
  }
};

// 启动应用
initApp();
