import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';

const API_URL = ''
//process.env.BACKEND_URL || 'http://localhost:8080'; -> replace at build time with actual backend URL if you needed

export default function App() {
    const [todos, setTodos] = useState([]);
    const [newTodo, setNewTodo] = useState('');
    const [health, setHealth] = useState('checking...');

    useEffect(() => {
        fetch(`${API_URL}/healthz`)
            .then(res => setHealth(res.ok ? 'healthy' : 'unhealthy'))
            .catch(() => setHealth('unreachable'));

        fetchTodos();
    }, []);

    const fetchTodos = async () => {
        try {
            const res = await fetch(`${API_URL}/todos`);
            if (!res.ok) {
                console.error('fetchTodos: bad response', res.status);
                setTodos([]);
                return;
            }
            const ct = (res.headers.get('content-type') || '').toLowerCase();
            let data = [];
            if (ct.includes('application/json')) {
                data = await res.json();
                if (!Array.isArray(data)) data = [];
            } else {
                // non-json -> fallback to empty list
                data = [];
            }
            setTodos(data);
        } catch (e) {
            console.error('Error fetching todos', e);
            setTodos([]);
        }
    };

    const addTodo = async () => {
        if (!newTodo.trim()) return;
        await fetch(`${API_URL}/todos`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ text: newTodo, done: false }),
        });
        setNewTodo('');
        fetchTodos();
    };

    const toggleTodo = async (id, done) => {
        await fetch(`${API_URL}/todos/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ done: !done }),
        });
        fetchTodos();
    };

    const deleteTodo = async (id) => {
        try {
            const res = await fetch(`${API_URL}/todos/${id}`, { method: 'DELETE' });
            if (!res.ok) console.error('deleteTodo failed', res.status);
        } catch (e) {
            console.error('deleteTodo error', e);
        } finally {
            await fetchTodos();
        }
    };

    return (
        <div className="min-h-screen bg-gray-100 flex flex-col items-center p-6">
            <div className="w-full max-w-md bg-white rounded-2xl shadow p-6">
                <h1 className="text-2xl font-bold mb-2">Todo List</h1>
                <p className="text-sm text-gray-500 mb-4">Backend health: <span className={health === 'healthy' ? 'text-green-600' : 'text-red-600'}>{health}</span></p>

                <div className="flex mb-4">
                    <input
                        value={newTodo}
                        onChange={(e) => setNewTodo(e.target.value)}
                        placeholder="Add new todo..."
                        className="flex-grow border border-gray-300 rounded-l-lg px-3 py-2 focus:outline-none"
                    />
                    <button
                        onClick={addTodo}
                        className="bg-blue-600 text-white px-4 rounded-r-lg hover:bg-blue-700"
                    >Add</button>
                </div>

                <ul>
                    {(Array.isArray(todos) ? todos : []).map((t) => (
                        <li key={t.id} className="flex justify-between items-center py-2 border-b border-gray-200">
                            <span
                                onClick={() => toggleTodo(t.id, t.done)}
                                className={`cursor-pointer ${t.done ? 'line-through text-gray-400' : ''}`}
                            >{t.text}</span>
                            <button
                                onClick={() => deleteTodo(t.id)}
                                className="text-red-500 hover:text-red-700"
                            >×</button>
                        </li>
                    ))}
                </ul>
            </div>
        </div>
    );
}

// Mount app to #root
if (typeof document !== 'undefined') {
    const el = document.getElementById('root');
    if (el) {
        const root = ReactDOM.createRoot(el);
        root.render(
            <React.StrictMode>
                <App />
            </React.StrictMode>
        );
    }
}
