import React, { useState, useEffect, useRef } from 'react';
import './App.css';
import { SetGoodPrice, GetAllBestProfit } from "../wailsjs/go/main/App";
import { Debug } from '../wailsjs/go/main/App';
import { SavePrice, RunCrawl} from '../wailsjs/go/main/App';

export default function App() {
    // 状态管理
    const [profits, setProfits] = useState([]);
    const [loading, setLoading] = useState(true);
    const [editingPrice, setEditingPrice] = useState(null); // 当前正在编辑的价格项
    const priceInputRef = useRef(null);
    const [errorMessage, setErrorMessage] = useState(''); // 错误信息
    const [showConfirmDialog, setShowConfirmDialog] = useState(false); // 确认对话框显示状态
    
    // 搜索相关状态
    const [searchTerm, setSearchTerm] = useState('');
    const [matchedIndices, setMatchedIndices] = useState([]);
    const [currentMatchIndex, setCurrentMatchIndex] = useState(0);
    const goodsRowRefs = useRef([]);

    // 加载利润数据
    useEffect(() => {
        async function fetchProfits() {
            try {
                const data = await GetAllBestProfit();
                setProfits(data);
                setLoading(false);
            } catch (error) {
                console.error("获取利润数据失败:", error);
                setLoading(false);
            }
        }
        
        fetchProfits();
    }, []);

    // 处理价格双击编辑
    const handlePriceDoubleClick = (profit, index) => {
        setEditingPrice({
            index,
            name: profit.name,
            value: profit.price || 0
        });
        // 使用setTimeout确保DOM元素已经渲染后再聚焦
        setTimeout(() => {
            if (priceInputRef.current) {
                priceInputRef.current.focus();
                priceInputRef.current.select();
            }
        }, 10);
    };

    // 处理价格编辑完成
    const handlePriceEditComplete = async (save = true) => {
        if (save && editingPrice) {
            try {
                const newPrice = parseInt(editingPrice.value, 10);
                if (!isNaN(newPrice) && newPrice >= 0) {
                    // 调用后端API更新价格
                    await SetGoodPrice(editingPrice.name, newPrice);
                    
                    // 更新本地状态
                    setProfits(prevProfits => {
                        const newProfits = [...prevProfits];
                        newProfits[editingPrice.index].price = newPrice;
                        return newProfits;
                    });
                    const data = await GetAllBestProfit();
                    setProfits(data);
                    setLoading(false);
                    SavePrice()
                }
            } catch (error) {
                console.error("更新价格失败:", error);
            }
        }
        setEditingPrice(null);
    };

    // 处理价格输入变化
    const handlePriceChange = (e) => {
        setEditingPrice(prev => ({
            ...prev,
            value: e.target.value
        }));
    };

    // 处理键盘事件
    const handleKeyDown = (e) => {
        if (e.key === 'Enter') {
            handlePriceEditComplete(true);
        } else if (e.key === 'Escape') {
            handlePriceEditComplete(false);
        }
    };

    // 格式化数字为带千位分隔符的字符串
    const formatNumber = (num) => {
        if (!num) return "0";
        
        // 处理小数和整数
        const parts = num.toString().split('.');
        parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ",");
        
        // 如果有小数部分，则添加回去
        return parts.length > 1 ? parts[0] : parts[0];
    };

    // 处理抓取数据按钮点击
    const handleCrawlClick = () => {
        // 显示确认对话框
        setShowConfirmDialog(true);
    };
    
    // 确认抓取数据
    const confirmCrawl = async () => {
        // 隐藏确认对话框
        setShowConfirmDialog(false);
        
        try {
            // 清除之前的错误信息
            setErrorMessage('');
            
            // 显示加载状态
            setLoading(true);
            
            // 调用抓取数据方法
            await RunCrawl();
            
            // 重新加载数据
            const data = await GetAllBestProfit();
            setProfits(data);
        } catch (error) {
            console.error("抓取数据失败:", error);
            // 设置错误信息
            setErrorMessage(error.toString());
        } finally {
            setLoading(false);
        }
    };
    
    // 取消抓取数据
    const cancelCrawl = () => {
        // 隐藏确认对话框
        setShowConfirmDialog(false);
    };
    
    // 处理搜索输入变化
    const handleSearchChange = (e) => {
        const value = e.target.value;
        setSearchTerm(value);
        
        if (value.trim() === '') {
            setMatchedIndices([]);
            setCurrentMatchIndex(0);
            return;
        }
        
        // 查找匹配项
        const matches = profits.reduce((acc, profit, index) => {
            if (profit.name.toLowerCase().includes(value.toLowerCase())) {
                acc.push(index);
            }
            return acc;
        }, []);
        
        setMatchedIndices(matches);
        setCurrentMatchIndex(matches.length > 0 ? 0 : -1);
        
        // 如果有匹配项，滚动到第一个匹配项
        if (matches.length > 0) {
            setTimeout(() => {
                scrollToMatch(matches[0]);
            }, 100);
        }
    };
    
    // 处理搜索框键盘事件
    const handleSearchKeyDown = (e) => {
        // 当按下 Enter 键时，跳转到下一个匹配项
        if (e.key === 'Enter') {
            e.preventDefault();
            goToNextMatch();
        }
    };
    
    // 跳转到下一个匹配项
    const goToNextMatch = () => {
        if (matchedIndices.length === 0) return;
        
        const nextIndex = (currentMatchIndex + 1) % matchedIndices.length;
        setCurrentMatchIndex(nextIndex);
        scrollToMatch(matchedIndices[nextIndex]);
    };
    
    // 滚动到指定匹配项
    const scrollToMatch = (rowIndex) => {
        if (goodsRowRefs.current[rowIndex]) {
            goodsRowRefs.current[rowIndex].scrollIntoView({
                behavior: 'smooth',
                block: 'center'
            });
        }
    };

    return (
        <div className="app-container">
            {/* 确认对话框 */}
            {showConfirmDialog && (
                <div className="confirm-dialog-overlay">
                    <div className="confirm-dialog">
                        <h3>抓取数据前请确认以下事项：</h3>
                        <ol>
                            <li>游戏必须保持窗口化和1080p分辨率。</li>
                            <li>界面应该处于交易中心界面，且左侧分栏未人为折叠、移动过。</li>
                            <li>应使用管理员权限启动本工具，确认后立刻自动切换到游戏界面进行采集。</li>
                            <li>大概2-3分钟到达【织造】后完成采集，可能会有数据错误，可以人工修正。</li>
                        </ol>
                        <div className="confirm-dialog-buttons">
                            <button
                                className="confirm-btn"
                                onClick={confirmCrawl}
                            >
                                确认
                            </button>
                            <button
                                className="cancel-btn"
                                onClick={cancelCrawl}
                            >
                                取消
                            </button>
                        </div>
                    </div>
                </div>
            )}
            
            {/* 浮动错误信息显示 */}
            {errorMessage && (
                <div className="floating-error-container">
                    <div className="error-message">
                        <span>{errorMessage}</span>
                        <button
                            className="close-error-btn"
                            onClick={() => setErrorMessage('')}
                            title="关闭错误信息"
                        >
                            ×
                        </button>
                    </div>
                </div>
            )}
            
            <div className="content-container">
                {/* 顶部工具栏 */}
                <div className="toolbar-container">
                    {/* 抓取数据按钮 */}
                    <button
                        className="crawl-btn"
                        onClick={handleCrawlClick}
                        title="抓取最新数据"
                        disabled={loading}
                    >
                        {loading ? "抓取中..." : "抓取数据"}
                    </button>
                    
                    {/* 搜索框 */}
                    <div className="search-container">
                        <input
                            type="text"
                            className="search-input"
                            placeholder="搜索货物名称... (按 Enter 跳转)"
                            value={searchTerm}
                            onChange={handleSearchChange}
                            onKeyDown={handleSearchKeyDown}
                        />
                        {matchedIndices.length > 0 && (
                            <div className="search-results">
                                <span className="match-count">
                                    {currentMatchIndex + 1}/{matchedIndices.length}
                                </span>
                                <button
                                    className="next-match-btn"
                                    onClick={goToNextMatch}
                                    title="查找下一个匹配项"
                                >
                                    ↓
                                </button>
                            </div>
                        )}
                    </div>
                </div>
                
                <div className="goods-list">
                    {loading ? (
                        <div className="loading">加载中...</div>
                    ) : (
                        <div className="goods-table">
                            <div className="goods-header">
                                <div className="goods-cell">名称</div>
                                <div className="goods-cell">参考价格</div>
                                <div className="goods-cell profit-cell">利润</div>
                                <div className="goods-cell comment-cell header-comment-cell">详情</div>
                            </div>
                            {profits.map((profit, index) => (
                                <div
                                    key={index}
                                    className={`goods-row ${matchedIndices.includes(index) ? 'matched-row' : ''} ${matchedIndices[currentMatchIndex] === index ? 'current-match' : ''}`}
                                    ref={el => goodsRowRefs.current[index] = el}
                                >
                                    <div className="goods-cell">{profit.name}</div>
                                    <div
                                        className="goods-cell price-cell"
                                        onDoubleClick={() => handlePriceDoubleClick(profit, index)}
                                    >
                                        {editingPrice && editingPrice.index === index ? (
                                            <input
                                                ref={priceInputRef}
                                                type="number"
                                                className="price-input"
                                                value={editingPrice.value}
                                                onChange={handlePriceChange}
                                                onBlur={() => handlePriceEditComplete(true)}
                                                onKeyDown={handleKeyDown}
                                                min="0"
                                            />
                                        ) : (
                                            <div className="price-display">
                                                {formatNumber(profit.price)}
                                                <span className="edit-hint">双击修改</span>
                                            </div>
                                        )}
                                    </div>
                                    <div className="goods-cell profit-cell">
                                        <span className={profit.num > 0 ? "profit-positive" : "profit-negative"}>
                                            {formatNumber(profit.num)}
                                        </span>
                                    </div>
                                    <div className="goods-cell comment-cell">
                                        <div className="comment-content">
                                            <span className="comment-preview">
                                                {profit.comment ? "..." : "-"}
                                            </span>
                                            {profit.comment && (
                                                <div className="comment-tooltip">
                                                    {profit.comment.split('\n').map((line, i) => (
                                                        <React.Fragment key={i}>
                                                            {line}
                                                            {i < profit.comment.split('\n').length - 1 && <br />}
                                                        </React.Fragment>
                                                    ))}
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}