import React, { useState, FormEvent, useEffect } from 'react';
import axios from 'axios';
import { toastifyMessage, MessageType, handleErrorWihtToast } from './Toastify.tsx';
import '../styles/TaskPage.css';
import Websocket from './Websocket.tsx';

export interface Task {
    id: string;
    title: string;
    description: string;
    status: string;
    estimated_time: string;
    actual_time: string;
    due_date: string;
    priority: string;
    workspace_id: string;
    assignee_id: string;
    created_at: string;
    updated_at: string;
    image_url: string;
}

interface Subtask {
    id: string;
    task_id: string;
    title: string;
    is_completed: string;
    assignee_id: string;
    created_at: string;
    updated_at: string;
}

interface Comment {
    id: string;
    task_id: string;
    user_id: string;
    text: string;
    username: string;
}

interface TaskPageProps {
    task: Task;
    workspace_id: string;
    onClose: () => void;
    onTaskUpdate: () => void;
}

const TaskPage: React.FC<TaskPageProps> = ({ task, workspace_id, onClose, onTaskUpdate }) => {
    const [updatedTask, setUpdatedTask] = useState<Task>({
        id: '',
        title: '',
        description: '',
        status: '',
        estimated_time: '',
        actual_time: '',
        due_date: '',
        priority: '',
        workspace_id: '',
        assignee_id: '',
        created_at: '',
        updated_at: '',
        image_url: '',
    });

    const websocketMessageHandler = (data: any) => {
        if (data.group === "subtask") {
            fetchSubtasks();
        } else if (data.group === "comment") {
            fetchComments();
        }

        if (data.type === "watch") {
            toastifyMessage(data.message, MessageType.Info);
        }
    };

    const [subtasks, setSubtasks] = useState<Subtask[]>([]);
    const [newSubtask, setNewSubtask] = useState({ title: '', assignee_id: '' });
    const [comments, setComments] = useState<Comment[]>([]);
    const [newComment, setNewComment] = useState('');
    const [isWatching, setIsWatching] = useState(false);

    const fetchSubtasks = async () => {
        try {
            const response = await axios.get(`/api/tasks/${task.id}/subtasks`);
            setSubtasks(response.data || []);
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to fetch subtasks.");
        }
    };

    const fetchComments = async () => {
        try {
            const response = await axios.get(`/api/workspaces/${workspace_id}/tasks/${task.id}/comments`);
            if (response.data != null && response.data.length !== 0) {
                const commentsWithUsernames = await Promise.all(response.data.map(async (comment: any) => {
                    const response2 = await axios.get(`/api/users/${comment.user_id}/profile/id`);
                    comment.username = response2.data.username;
                    return comment;
                }));
                setComments(commentsWithUsernames || []);
            }
            else {
                setComments([]);
            }
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to fetch comments.");
        }
    };

    const fetchWatchStatus = async () => {
        try {
            const response = await axios.get(`/api/workspaces/${workspace_id}/tasks/${task.id}/watch`);
            setIsWatching(response.data.status === 'Yes');
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to fetch watch status.");
        }
    };

    useEffect(() => {
        fetchSubtasks();
        fetchComments();
        fetchWatchStatus();
        Websocket.close();
        Websocket.connect(websocketMessageHandler)
    }, [task.id]);

    const handleTaskUpdate = async (event: FormEvent<HTMLFormElement>, updated: any) => {
        event.preventDefault();

        try {
            if (updated.title === '') {
                updated.title = task.title;
            }
            if (updated.description === '') {
                updated.description = task.description;
            }
            if (updated.actual_time === '') {
                updated.actual_time = task.actual_time;
            }
            if (updated.due_date === '') {
                updated.due_date = task.due_date;
            }
            if (updated.priority === '') {
                updated.priority = task.priority;
            }
            if (updated.assignee_id === '') {
                updated.assignee_id = task.assignee_id;
            }
            if (updated.image_url === '') {
                updated.image_url = task.image_url;
            }
            await axios.put(`/api/workspaces/${workspace_id}/tasks/${task.id}`, {
                title: updated.title,
                description: updated.description,
                actual_time: updated.actual_time,
                due_date: updated.due_date,
                priority: updated.priority,
                assignee_id: updated.assignee_id,
                image_url: updated.image_url
            });
            setUpdatedTask(updated);
            onTaskUpdate();
            toastifyMessage(`Task updated successfully`, MessageType.Success);
            onClose();
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to update task.");
        }
        setUpdatedTask({
            id: '',
            title: '',
            description: '',
            status: '',
            estimated_time: '',
            actual_time: '',
            due_date: '',
            priority: '',
            workspace_id: '',
            assignee_id: '',
            created_at: '',
            updated_at: '',
            image_url: '',
        });
    };

    const handleSubtaskDelete = async (subtaskId: string) => {
        try {
            await axios.delete(`/api/tasks/${task.id}/subtasks/${subtaskId}`);
            setSubtasks(subtasks.filter(subtask => subtask.id !== subtaskId));
            toastifyMessage('Subtask deleted successfully', MessageType.Success);
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to delete subtask.");
        }
    };

    const handleSubtaskAdd = async (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();

        try {
            const response = await axios.post(`/api/tasks/${task.id}/subtasks`, newSubtask);
            setSubtasks([...subtasks, response.data]);
            setNewSubtask({ title: '', assignee_id: '' });
            toastifyMessage('Subtask added successfully', MessageType.Success);
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to add subtask.");
        }
    };

    const calculateProgress = () => {
        const completedSubtasks = subtasks.filter(subtask => subtask.is_completed === 'Yes');
        return Math.round((completedSubtasks.length / subtasks.length) * 10000) / 100;
    };

    const handleSubtaskIsCompletedUpdate = async (subtask: Subtask) => {
        try {
            await axios.put(`/api/tasks/${task.id}/subtasks/${subtask.id}/status`, subtask);
            setSubtasks(subtasks.map(st => (st.id === subtask.id ? subtask : st)));
            toastifyMessage('Subtask updated successfully', MessageType.Success);
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to update subtask.");
        }
    };

    const handleSubtaskTitleUpdate = async (subtask: Subtask) => {
        try {
            await axios.put(`/api/tasks/${task.id}/subtasks/${subtask.id}/title`, subtask);
            setSubtasks(subtasks.map(st => (st.id === subtask.id ? subtask : st)));
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to update subtask.");
        }
    };

    const handleSubtaskAssigneeIDUpdate = async (subtask: Subtask) => {
        try {
            await axios.put(`/api/tasks/${task.id}/subtasks/${subtask.id}/assigneeid`, subtask);
            setSubtasks(subtasks.map(st => (st.id === subtask.id ? subtask : st)));
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to update subtask.");
        }
    };

    const handleCommentAdd = async (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();

        try {
            const response = await axios.post(`/api/workspaces/${workspace_id}/tasks/${task.id}/comments`, { text: newComment });
            const response2 = await axios.get(`/api/users/${response.data.user_id}/profile/id`)
            const user_username = response2.data.username;
            setComments([...comments, { ...response.data, username: user_username }]);
            setNewComment('');
            toastifyMessage('Comment added successfully', MessageType.Success);
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to add comment.");
        }
    };

    const handleWatchToggle = async () => {
        try {
            if (isWatching) {
                await axios.delete(`/api/workspaces/${workspace_id}/tasks/${task.id}/watch`);
                toastifyMessage('Stopped watching task', MessageType.Success);
            } else {
                await axios.post(`/api/workspaces/${workspace_id}/tasks/${task.id}/watch`);
                toastifyMessage('Started watching task', MessageType.Success);
            }
            setIsWatching(!isWatching);
        } catch (error: any) {
            handleErrorWihtToast(error, "Failed to update watch status.");
        }
    };

    return (
        <div className="taskpage-modal">
            <div className="taskpage-content">
                <div className="task-info">
                    <h2>Task Details</h2>
                    <p><strong>Title:</strong> {task.title}</p>
                    <p><strong>Actual Time:</strong> {task.actual_time}</p>
                    <p><strong>Due Date:</strong> {task.due_date}</p>
                    <p><strong>Estimated Time:</strong> {task.estimated_time}</p>
                    <p><strong>Priority:</strong> {task.priority}</p>
                    <p><strong>Assignee ID:</strong> {task.assignee_id}</p>
                    <p><strong>Created At:</strong> {task.created_at}</p>
                    <p><strong>Updated At:</strong> {task.updated_at}</p>
                    <p><strong>Description:</strong> {task.description}</p>
                    <div className="watch-toggle">
                        <button
                            onClick={handleWatchToggle}
                            className={!isWatching ? 'watch-btn' : 'unwatch-btn'}
                        >
                            {isWatching ? 'Unwatch' : 'Watch'}
                        </button>
                    </div>
                </div>
                <div className="subtask-comment-list">
                    <div className="subtask-list">
                        <h3>Subtasks</h3>
                        {subtasks != null && subtasks.length > 0 && (
                            <div>
                                <p>Progress:{calculateProgress()}%</p>
                                <div className="progress-bar">
                                    <div
                                        className="progress-bar-fill"
                                        style={{ width: `${calculateProgress()}%` }}
                                    ></div>
                                </div>
                            </div>
                        )}
                        {subtasks != null && subtasks.length > 0 ? (
                            subtasks.map((subtask) => (
                                <div key={subtask.id} className="subtask-item">
                                    <input
                                        type="checkbox"
                                        checked={subtask.is_completed === "Yes"}
                                        onChange={() => handleSubtaskIsCompletedUpdate({ ...subtask, is_completed: (subtask.is_completed == "Yes" ? "No" : "Yes") })}
                                    />
                                    <input
                                        type="number"
                                        value={subtask.assignee_id}
                                        onChange={(e) => handleSubtaskAssigneeIDUpdate({ ...subtask, assignee_id: e.target.value })}
                                    />
                                    <input
                                        type="text"
                                        value={subtask.title}
                                        onChange={(e) => handleSubtaskTitleUpdate({ ...subtask, title: e.target.value })}
                                    />
                                    <button onClick={() => handleSubtaskDelete(subtask.id)}>Delete</button>
                                </div>
                            ))
                        ) : (
                            <p>No subtasks available.</p>
                        )}
                        <form onSubmit={handleSubtaskAdd} className="add-subtask-form">
                            <input
                                type="text"
                                placeholder="New subtask title"
                                value={newSubtask.title}
                                onChange={(e) => setNewSubtask({ ...newSubtask, title: e.target.value })}
                            />
                            <input
                                type="number"
                                placeholder="Assignee ID"
                                value={newSubtask.assignee_id}
                                onChange={(e) => setNewSubtask({ ...newSubtask, assignee_id: e.target.value })}
                            />
                            <button type="submit">Add</button>
                        </form>
                    </div>
                    <div className="comment-list">
                        <h3>Comments</h3>
                        {comments != null && comments.length > 0 ? (
                            comments.map((comment) => (
                                <div key={comment.id}>{comment.username}: {comment.text}</div>
                            ))
                        ) : (
                            <p>No comments available.</p>
                        )}
                        <form className="add-comment-form" onSubmit={handleCommentAdd}>
                            <h3>Add Comment</h3>
                            <textarea
                                placeholder="Write a comment"
                                value={newComment}
                                onChange={(e) => setNewComment(e.target.value)}
                            />
                            <button type="submit">Add Comment</button>
                        </form>
                    </div>
                </div>
                <div className="task-update-form">
                    <form onSubmit={(event) => handleTaskUpdate(event, updatedTask)}>
                        <label>Title</label>
                        <input
                            type="text"
                            value={updatedTask.title}
                            onChange={(event) => setUpdatedTask({ ...updatedTask, title: event.target.value })}
                        />
                        <label>Description</label>
                        <textarea
                            value={updatedTask.description}
                            onChange={(event) => setUpdatedTask({ ...updatedTask, description: event.target.value })}
                        ></textarea>
                        <label>Actual Time</label>
                        <input
                            type="number"
                            value={updatedTask.actual_time}
                            onChange={(event) => setUpdatedTask({ ...updatedTask, actual_time: event.target.value })}
                        ></input>
                        <label>Due Date</label>
                        <input
                            type="date"
                            value={updatedTask.due_date}
                            onChange={(event) => setUpdatedTask({ ...updatedTask, due_date: event.target.value })}
                        />
                        <label>Assignee ID</label>
                        <input
                            type="number"
                            value={updatedTask.assignee_id}
                            onChange={(event) => setUpdatedTask({ ...updatedTask, assignee_id: event.target.value })}
                        />
                        <label>Image URL</label>
                        <input
                            type="text"
                            value={updatedTask.image_url}
                            onChange={(event) => setUpdatedTask({ ...updatedTask, image_url: event.target.value })}
                        />
                        <div className="button-group">
                            <button type="submit" className='update-button'>Update</button>
                            <button type="button" className="close-button" onClick={onClose}>Close</button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    );
}

export default TaskPage;
