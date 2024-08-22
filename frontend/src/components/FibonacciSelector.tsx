import React from 'react';
import '../styles/FibonacciSelector.css'

interface FibonacciSelectorProps {
    selectedValue: number;
    onSelect: (value: number) => void;
}

const FibonacciSelector: React.FC<FibonacciSelectorProps> = ({ selectedValue, onSelect }) => {
    const fibonacciNumbers = [1, 2, 3, 5, 8, 13, 21];

    return (
        <div className="fibonacci-selector">
            <div className="selected-value">{selectedValue}</div>
            <div className="buttons">
                {fibonacciNumbers.map(number => (
                    <button
                        key={number}
                        className={selectedValue === number ? 'selected' : ''}
                        onClick={() => onSelect(number)}
                    >
                        {number}
                    </button>
                ))}
            </div>
        </div>
    );
};

export default FibonacciSelector;
