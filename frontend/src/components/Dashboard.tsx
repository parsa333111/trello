import { useState, useEffect, FormEvent } from 'react';
import axios from 'axios';
import { toastifyMessage, MessageType, handleErrorWihtToast } from './Toastify.tsx';
import { useNavigate } from 'react-router-dom';
import Websocket from './Websocket'
import '../styles/Dashboard.css';
import logo_bw_stable from '/logo-bw-stable.gif';
import logo_bw_animate from '/logo-bw-animate.gif';
import settings_icon from '/settings-icon.png';
import { Task } from './TaskPage.tsx';

interface Workspace {
    id: string;
    name: string;
    description: string;
    created_at: string;
    updated_at: string;
}

const Dashboard = () => {
    const [isSettingPopupVisible, setIsSettingPopupVisible] = useState(false);
    const [isAccountPopupVisible, setIsAccountPopupVisible] = useState(false);
    const [isWorkspaceFormVisible, setIsWorkspaceFormVisible] = useState(false);
    const [logo, setLogo] = useState(logo_bw_stable);
    const [userEmail, setUserEmail] = useState('');
    const [userUsername, setUserUsername] = useState('');

    const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
    const [tasks, setTasks] = useState<Task[]>([]);

    const [workspaceMembers, setWorkspaceMembers] = useState<Map<string, number>>(new Map());
    const [workspaceMember, setWorkspaceMember] = useState<string>('');

    const [workspacesPage, setWorkspacesPage] = useState(1);
    const [tasksPage, setTasksPage] = useState(1);

    const [workspacesPerPage] = useState(4);
    const [tasksPerPage] = useState(3);

    const getPaginatedItems = (data: any[], page: number, perPage: number) => {
        const startIndex = (page - 1) * perPage;
        const endIndex = startIndex + perPage;
        return data.slice(startIndex, endIndex);
    };

    const websocketMessageHandler = (data: any) => {
        if (data.group === "workspace") {
            fetchWorkspaces();
        }
        else if (data.group === "task") {
            fetchTasks();
        }

        if (data.type === "watch") {
            toastifyMessage(data.message, MessageType.Info);
        }
    }

    const navigate = useNavigate();

    const fetchUserProfile = async () => {
        try {
            const response = await axios.get('/api/users/self/profile');
            setUserEmail(response.data.email);
            setUserUsername(response.data.username);
        } catch (error: any) {
            handleErrorWihtToast(error, 'Failed to fetch profile.');
        }
    };

    const fetchWorkspaces = async () => {
        try {
            const response = await axios.get('/api/workspaces');
            setWorkspaces(response.data);
        } catch (error: any) {
            console.error('Failed to fetch workspaces:', error);
        }
    };

    const fetchTasks = async () => {
        try {
            const response = await axios.get('/api/self/tasks');
            setTasks(response.data);
        } catch (error: any) {
            handleErrorWihtToast(error, 'Failed to fetch tasks.');
        }
    };

    useEffect(() => {
        fetchWorkspaces();
        fetchTasks();
        Websocket.close();
        Websocket.connect(websocketMessageHandler);
    }, []);

    useEffect(() => {
        if (isSettingPopupVisible) {
            fetchUserProfile();
        }
    }, [isSettingPopupVisible]);

    const handleSettingClick = () => {
        setIsSettingPopupVisible(!isSettingPopupVisible);
    };

    const handleAccountClick = () => {
        setIsAccountPopupVisible(true);
    };

    const handleCloseAccountPopup = () => {
        setIsAccountPopupVisible(false);
    };

    const handleLogout = () => {
        localStorage.removeItem('accessToken');
        localStorage.removeItem('refreshToken');
        Websocket.close();

        toastifyMessage("Logged out successfully.", MessageType.Success);

        navigate('/login');
    };

    const handleDeleteAccount = async () => {
        const confirmed = window.confirm("Are you sure you want to delete your account? This action cannot be undone.");

        if (confirmed) {
            try {
                await axios.delete('/api/users/self/profile');
                toastifyMessage("Account has been deleted.", MessageType.Warning);
                setIsAccountPopupVisible(false);
                handleLogout();
            } catch (error: any) {
                handleErrorWihtToast(error, 'Failed to delete account.');
            }
        } else {
            toastifyMessage("Account deletion cancelled.", MessageType.Warning);
        }
    };

    const handleUpdateUsername = async (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        const newUsername = (event.target as HTMLFormElement).elements.namedItem('newUsername') as HTMLInputElement;
        try {
            await axios.put('/api/users/self/profile/username', { username: newUsername.value });
            setUserUsername(newUsername.value);
            setIsAccountPopupVisible(false);
        } catch (error: any) {
            handleErrorWihtToast(error, 'Failed to update username.');
        }
    };

    const handleUpdatePassword = async (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        const newPassword = (event.target as HTMLFormElement).elements.namedItem('newPassword') as HTMLInputElement;
        try {
            await axios.put('/api/users/self/profile/password', { password: newPassword.value });
            setIsAccountPopupVisible(false);
        } catch (error: any) {
            handleErrorWihtToast(error, 'Failed to update password.');
        }
    };

    const handleCreateWorkspaceClick = () => {
        setIsWorkspaceFormVisible(true);
    };

    const handleCreateWorkspace = async (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();

        const name = (event.target as HTMLFormElement).elements.namedItem('name') as HTMLInputElement;
        const description = (event.target as HTMLFormElement).elements.namedItem('description') as HTMLInputElement;

        try {
            const response = await axios.post('/api/workspaces', { name: name.value, description: description.value });
            const workspace_id = response.data.id;

            for (const member of workspaceMembers) {
                try {
                    await axios.post(`/api/workspaces/${workspace_id}/users`, { user_id: member[1], role: "StandardUser" });
                } catch (error: any) {
                    handleErrorWihtToast(error, `Failed to add '${member[0]}' to the workspace.`);
                }
            }

            fetchWorkspaces();
        } catch (error: any) {
            handleErrorWihtToast(error, 'Failed to create workspace.');
        }

        setIsWorkspaceFormVisible(false);
        setWorkspaceMembers(new Map())
    };

    const handleAddMember = async () => {
        try {
            const response = await axios.get(`/api/users/${workspaceMember}/profile/username`)
            const user_id = response.data.id;

            if (workspaceMember.trim() !== '') {
                setWorkspaceMembers((prev) => {
                    const updated = new Map(prev);
                    updated.set(workspaceMember, user_id);
                    return updated;
                });
                setWorkspaceMember('');
            }
            else {
                toastifyMessage("Invalid username format.", MessageType.Error)
            }
        }
        catch (error: any) {
            handleErrorWihtToast(error, 'Invalid username for the new memeber.');
        }
    };

    const handleRemoveMember = (member: string) => {
        setWorkspaceMembers((prev) => {
            const updated = new Map(prev);
            updated.delete(member);
            return updated;
        });
    };

    const handleLogoMouseEnter = () => {
        setLogo(logo_bw_animate);
    };

    const handleLogoMouseLeave = () => {
        setLogo(logo_bw_stable);
    };

    const getRandomColor = (seed: any) => {
        const letters = '0123456789ABCDEF';
        let color = '#';
        let random = 0.1111;
        for (let i = 0; i < seed; i++) {
            random *= 2;
            if (random > 1) {
                random -= 1;
            }
        }
        for (let i = 0; i < 6; i++) {
            color += letters[Math.floor(random * 16)];
            random *= 2;
            if (random > 1) {
                random -= 1;
            }
        }
        return color;
    };

    const handleWorkspaceClick = (workspace_id: string) => {
        Websocket.close();
        navigate(`/workspace/${workspace_id}`);
    };

    const handleTaskClick = (workspace_id: string) => {
        Websocket.close();
        navigate(`/workspace/${workspace_id}`);
    };

    return (
        <div className="dashboard">
            <header className="header">
                <div className="left-container">
                    <div className="logo-container">
                        <img
                            src={logo}
                            alt="Trello Logo"
                            onMouseEnter={handleLogoMouseEnter}
                            onMouseLeave={handleLogoMouseLeave}
                        />
                    </div>
                    <button className="create-btn" onClick={handleCreateWorkspaceClick}>Create</button>
                </div>
                <div className="right-container">
                    <div className="settings-container" onMouseEnter={handleSettingClick} onMouseLeave={handleSettingClick}>
                        <img src={settings_icon} alt="Settings" />
                        {isSettingPopupVisible && (
                            <div className="popup">
                                <div className="popup-row" onClick={handleAccountClick}>
                                    <div>{userUsername}</div>
                                    <div>{userEmail}</div>
                                </div>
                                <div className="popup-row" onClick={handleLogout}>Logout</div>
                                <div className="popup-row delete-account" onClick={handleDeleteAccount}>Delete Account</div>
                            </div>
                        )}
                    </div>
                </div>
            </header>
            <div className="content">
                <div className="workspaces">
                    <h2>Your Workspaces</h2>
                    {workspaces != null && workspaces.length > 0 ? (
                        <>
                            <div className="pagination">
                                <button
                                    className="previous-page-btn"
                                    onClick={() => setWorkspacesPage(workspacesPage - 1)} disabled={workspacesPage === 1}
                                >
                                    Previous
                                </button>
                                <span> <strong>Page {workspacesPage}</strong> </span>
                                <button
                                    className="next-page-btn"
                                    onClick={() => setWorkspacesPage(workspacesPage + 1)}
                                    disabled={workspaces.length <= workspacesPage * workspacesPerPage}
                                >
                                    Next
                                </button>
                            </div>
                            {getPaginatedItems(workspaces, workspacesPage, workspacesPerPage).map((workspace) => (
                                <div
                                    key={workspace.id}
                                    className="workspace"
                                    style={{ backgroundColor: getRandomColor(workspace.id) }}
                                    onClick={() => handleWorkspaceClick(workspace.id)}
                                >
                                    <div className="workspace-title">
                                        <span>{workspace.name}</span>
                                    </div>
                                    <div className="workspace-description">
                                        <span><strong>Description:</strong> {workspace.description}</span>
                                    </div>
                                </div>
                            ))}
                        </>
                    ) : (
                        <p>No workspaces are available.</p>
                    )}
                </div>
                <div className="tasks">
                    <h2>Your Tasks</h2>
                    {tasks != null && tasks.length > 0 ? (
                        <>
                            <div className="pagination">
                                <button
                                    className="previous-page-btn"
                                    onClick={() => setTasksPage(tasksPage - 1)} disabled={tasksPage === 1}
                                >
                                    Previous
                                </button>
                                <span> <strong>Page {tasksPage}</strong> </span>
                                <button
                                    className="next-page-btn"
                                    onClick={() => setTasksPage(tasksPage + 1)}
                                    disabled={tasks.length <= tasksPage * tasksPerPage}
                                >
                                    Next
                                </button>
                            </div>
                            {getPaginatedItems(tasks, tasksPage, tasksPerPage).map((task) => (
                                <div key={task.id} className="task" onClick={() => handleTaskClick(task.workspace_id)}>
                                    <h2>{task.title}</h2>
                                    <p><strong>Due Date:</strong> {task.due_date}</p>
                                    <p><strong>Priority:</strong> {task.priority}</p>
                                    <p><strong>Status:</strong> {task.status}</p>
                                </div>
                            ))}
                        </>
                    ) : (
                        <p>No tasks are available.</p>
                    )}
                </div>
            </div>
            {isAccountPopupVisible && (
                <div className="modal-overlay" onClick={handleCloseAccountPopup}>
                    <div className="modal" onClick={(e) => e.stopPropagation()}>
                        <h2>Update Account</h2>
                        <form onSubmit={handleUpdateUsername}>
                            <div className="form-group">
                                <label>New Username:</label>
                                <input type="text" name="newUsername" required />
                            </div>
                            <div className="form-actions">
                                <button type="submit" className="create-btn">Update Username</button>
                            </div>
                        </form>
                        <form onSubmit={handleUpdatePassword}>
                            <div className="form-group">
                                <label>New Password:</label>
                                <input type="text" name="newPassword" required />
                            </div>
                            <div className="form-actions">
                                <button type="submit" className="create-btn">Update Password</button>
                                <button type="button" className="cancel-btn" onClick={handleCloseAccountPopup}>Cancel</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
            {isWorkspaceFormVisible && (
                <div className="modal-overlay" onClick={() => setIsWorkspaceFormVisible(false)}>
                    <div className="modal" onClick={(e) => e.stopPropagation()}>
                        <h2>Create New Workspace</h2>
                        <form onSubmit={handleCreateWorkspace}>
                            <div className="form-group">
                                <label>Name:</label>
                                <input type="text" name="name" required />
                            </div>
                            <div className="form-group">
                                <label>Description:</label>
                                <input type="text" name="description" required />
                            </div>
                            <div className="form-group">
                                <label>Workspace Members:</label>
                                <div style={{ display: 'flex', alignItems: 'center', marginBottom: '10px' }}>
                                    <input
                                        type="text"
                                        value={workspaceMember}
                                        onChange={(e) => setWorkspaceMember(e.target.value)}
                                        style={{ marginRight: '10px' }}
                                    />
                                    <button type="button" className="add-btn" onClick={handleAddMember}>Add</button>
                                </div>
                            </div>
                            <div className="form-group">
                                <ul className="workspace-members-list">
                                    {Array.from(workspaceMembers.entries()).map((member, index) => (
                                        <li key={index} className="workspace-members-list-row">
                                            {member[0]}
                                            <button type="button" className="remove-btn" onClick={() => handleRemoveMember(member[0])}>
                                                remove
                                            </button>
                                        </li>
                                    ))}
                                </ul>
                            </div>
                            <div className="form-actions">
                                <button type="button" className="cancel-btn" onClick={() => setIsWorkspaceFormVisible(false)}>Cancel</button>
                                <button type="submit" className="create-btn">Create Workspace</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Dashboard;
