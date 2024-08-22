import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useNavigate, useParams } from 'react-router-dom';
import {
    DndContext,
    closestCenter,
    KeyboardSensor,
    PointerSensor,
    useSensor,
    useSensors,
    DragEndEvent
} from '@dnd-kit/core';
import {
    SortableContext,
    sortableKeyboardCoordinates,
    useSortable,
    verticalListSortingStrategy
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import '../styles/Workspace.css';
import logo_bw_stable from '/logo-bw-stable.gif';
import logo_bw_animate from '/logo-bw-animate.gif';
import online_icon from '/online-icon.png';
import offline_icon from '/offline-icon.png';
import FibonacciSelector from './FibonacciSelector';
import TaskPage, { Task } from './TaskPage.tsx';
import { toastifyMessage, MessageType, handleErrorWihtToast } from './Toastify.tsx';
import MemberDetails, { MemberUtils } from './MemberDetails.tsx'
import Websocket from './Websocket.tsx';

const Workspace = () => {
    const [logo, setLogo] = useState(logo_bw_stable);

    const [taskPlanned, setTaskPlanned] = useState<Task[]>([]);
    const [taskInProgress, setTaskInProgress] = useState<Task[]>([]);
    const [taskCompleted, setTaskCompleted] = useState<Task[]>([]);
    const [selectedTask, setSelectedTask] = useState<Task | null>(null);

    const [showInputPlanned, setShowInputPlanned] = useState(false);
    const [isTaskPageVisible, setTaskPageVisible] = useState(false);

    const [membersUtils, setMembersUtils] = useState<MemberUtils[]>([]);
    const [selectedMemberUtils, setSelectedMemberUtils] = useState<MemberUtils | null>(null);
    const [isMemberDetailsVisible, setMemberDetailsVisible] = useState(false);

    const [file, setFile] = useState(null);

    const [newTask, setNewTask] = useState<Task>({
        id: '',
        title: '',
        description: '',
        status: '',
        estimated_time: '1',
        actual_time: '',
        due_date: '',
        priority: '',
        workspace_id: '',
        assignee_id: '',
        created_at: '',
        updated_at: '',
        image_url: '',
    });

    const navigate = useNavigate();
    const { workspace_id } = useParams<{ workspace_id: string }>();

    const websocketMessageHandler = (data: any) => {
        if (data.group === "task") {
            fetchTasks();
        } else if (data.group === "memeber") {
            fetchMembers();
        }

        if (data.type === "watch") {
            toastifyMessage(data.message, MessageType.Info);
        }
    };

    const fetchTasks = async () => {
        try {
            const response = await axios.get(`/api/workspaces/${workspace_id}/tasks`);
            const tasks = response.data || [];

            setTaskPlanned(tasks.filter((task: Task) => task.status === 'Planned'));
            setTaskInProgress(tasks.filter((task: Task) => task.status === 'InProgress'));
            setTaskCompleted(tasks.filter((task: Task) => task.status === 'Completed'));
        } catch (error: any) {
            handleErrorWihtToast(error, 'Failed to fetch tasks.');
        }
    };

    const fetchMembers = async () => {
        try {
            const response = await axios.get(`/api/workspaces/${workspace_id}/users`);
            const members = response.data || [];

            const members_utils: MemberUtils[] = []

            for (const member of members) {
                const response = await axios.get(`/api/users/${member.user_id}/profile/id`);
                const member_profile = response.data;

                members_utils.push(
                    {
                        member: member,
                        member_profile: member_profile,
                    }
                )
            }

            setMembersUtils(members_utils);
        } catch (error: any) {
            handleErrorWihtToast(error, 'Failed to fetch members.');
        }
    };

    useEffect(() => {
        fetchTasks();
        fetchMembers();
        Websocket.close();
        Websocket.connect(websocketMessageHandler);
    }, [workspace_id]);

    const handleLogoMouseEnter = () => {
        setLogo(logo_bw_animate);
    };

    const handleLogoMouseLeave = () => {
        setLogo(logo_bw_stable);
    };

    const handleTaskClick = (task: Task) => {
        setTaskPageVisible(true);
        setSelectedTask(task);
    };

    const handleMemberClick = (member_utils: MemberUtils) => {
        setMemberDetailsVisible(true);
        setSelectedMemberUtils(member_utils);
    };

    const handleAddCardClick = (list: string) => {
        if (list === 'Planned') {
            setShowInputPlanned(true);
        }
        setNewTask({ ...newTask, status: list });
    };

    const handleCancelAddCard = (list: string) => {
        if (list === 'Planned') {
            setShowInputPlanned(false);
        }
        resetNewTask();
    };

    const handleNewTaskChange = (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        const { name, value } = event.target;
        setNewTask(prevState => ({ ...prevState, [name]: value }));
    };

    const handleExitWorkspace = () => {
        Websocket.close();
        navigate('/dashboard');
    };

    const handleLeaveWorkspace = async () => {
        const confirmed = window.confirm("Are you sure you want to leave this workspace? This action cannot be undone.");

        if (confirmed) {
            try {
                await axios.delete(`/api/workspaces/${workspace_id}/users/leave`);
                toastifyMessage("You have left the workspace.", MessageType.Warning);
                handleExitWorkspace();
            } catch (error: any) {
                handleErrorWihtToast(error, 'Failed to the leave workspace.');
            }
        } else {
            toastifyMessage("Leaving operation cancelled.", MessageType.Warning);
        }
    };

    const handleDeleteWorkspace = async () => {
        const confirmed = window.confirm("Are you sure you want to delete this workspace? This action cannot be undone.");

        if (confirmed) {
            try {
                await axios.delete(`/api/workspaces/${workspace_id}`);
                toastifyMessage("Workspace has been deleted.", MessageType.Warning);
                handleExitWorkspace();
            } catch (error: any) {
                handleErrorWihtToast(error, 'Failed to delete the workspace.');
            }
        } else {
            toastifyMessage("Workspace deletion cancelled.", MessageType.Warning);
        }
    };

    const resetNewTask = () => {
        setNewTask({
            id: '',
            title: '',
            description: '',
            status: '',
            estimated_time: '1',
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

    const handleAddCard = async (list: string) => {
        try {
            if (newTask.description === '') {
                newTask.description = 'None';
            }
            if (newTask.actual_time === '') {
                newTask.actual_time = '1';
            }
            if (newTask.priority === '') {
                newTask.priority = '1';
            }
            if (newTask.image_url === '') {
                newTask.image_url = 'No Image';
            }
            const response = await axios.post(`/api/workspaces/${workspace_id}/tasks`, {
                title: newTask.title,
                description: newTask.description,
                estimated_time: newTask.estimated_time,
                actual_time: newTask.actual_time,
                due_date: newTask.due_date,
                priority: newTask.priority,
                assignee_id: newTask.assignee_id,
                image_url: newTask.image_url,
            });
            console.log(file)
            fileUpload(response.data.id)
            resetNewTask();
            fetchTasks();
            handleCancelAddCard(list);
        } catch (error: any) {
            handleErrorWihtToast(error, 'Failed to add task.');
            if (newTask.description === 'None') {
                newTask.description = '';
            }
            if (newTask.actual_time === '1') {
                newTask.actual_time = '';
            }
            if (newTask.priority === '1') {
                newTask.priority = '';
            }
            if (newTask.image_url === 'No Image') {
                newTask.image_url = '';
            }
        }
    };

    const handleOnDragEnd = async (event: DragEndEvent) => {
        const { active, over } = event;

        if (!over) return;

        const sourceContainerId: string = active.data.current?.sortable.containerId;
        const destinationContainerId: string = over.data.current?.sortable.containerId;

        if (sourceContainerId !== destinationContainerId) {
            const updatedTasks = new Map<string, Task[]>();
            updatedTasks.set('Planned', [...taskPlanned]);
            updatedTasks.set('InProgress', [...taskInProgress]);
            updatedTasks.set('Completed', [...taskCompleted]);

            const draggedTaskId = active.id;
            const sourceTasks = updatedTasks.get(sourceContainerId) || [];
            const destinationTasks = updatedTasks.get(destinationContainerId) || [];

            const draggedTask = sourceTasks.find(task => task.id === draggedTaskId);

            if (draggedTask) {
                try {
                    await axios.put(`/api/workspaces/${workspace_id}/tasks/${draggedTask.id}/status`, {
                        status: destinationContainerId
                    });

                    const draggedIndex = sourceTasks.findIndex(task => task.id === draggedTaskId);
                    sourceTasks.splice(draggedIndex, 1);

                    draggedTask.status = destinationContainerId;
                    destinationTasks.push(draggedTask);

                    updatedTasks.set(sourceContainerId, sourceTasks);
                    updatedTasks.set(destinationContainerId, destinationTasks);

                    setTaskPlanned(updatedTasks.get('Planned') || []);
                    setTaskInProgress(updatedTasks.get('InProgress') || []);
                    setTaskCompleted(updatedTasks.get('Completed') || []);

                } catch (error: any) {
                    handleErrorWihtToast(error, 'Failed to update task status');
                }
            }
        }
    };


    const handleFibonacciSelect = (value: number) => {
        setNewTask(prevState => ({ ...prevState, estimated_time: value.toString() }));
    };

    const SortableItem = (props: { task: Task }) => {
        const { attributes, listeners, setNodeRef, transform, transition } = useSortable({ id: props.task.id });
        const style = {
            transform: CSS.Transform.toString(transform),
            transition,
        };
    
        const [imageSrc, setImageSrc] = useState('');
    
        useEffect(() => {
            const fetchImage = async () => {
                try {
                    const response = await fetch(`/api/retrieve/picture/${props.task.id}`);
                    if (response.ok && response.status != 204) {
                        const blob = await response.blob();
                        const url = URL.createObjectURL(blob);
                        setImageSrc(url);
                    } else {
                        setImageSrc('');
                    }
                } catch (error) {
                    setImageSrc('');
                }
            };
    
            if (props.task.id) {
                fetchImage();
            }
        }, [props.task.id]);
    
        return (
            <div
                ref={setNodeRef}
                style={style}
                {...attributes}
                {...listeners}
                className="task-card"
                onClick={() => handleTaskClick(props.task)}
            >
                <h3>{props.task.title}</h3>
                <p>{props.task.description}</p>
                {imageSrc != '' && <img src={imageSrc} alt="Fetched" />}
            </div>
        );
    };
    

    const PlaceHolder = () => {
        const { attributes, listeners, setNodeRef, transform, transition } = useSortable({ id: "empty" });
        const style = {
            transform: CSS.Transform.toString(transform),
            transition,
        };

        return (
            <div
                ref={setNodeRef}
                style={style}
                {...attributes}
                {...listeners}
            >
                <p>Looks like it's empty!</p>
            </div>
        );
    }

    const onFileChange = (e: any) => {
        setFile(e.target.files[0]);
    };

    const fileUpload = async (id: any) => {

        if (!file) {
            return;
        }

        const formData = new FormData();
        formData.append('file', file);

        try {
            await axios.post(`/api/upload/picture/${id}`, formData, {
                headers: {
                    'Content-Type': 'multipart/form-data',
                },
            });
        } catch (err: any) {
            handleErrorWihtToast(err, 'File upload failed');
        }
    };

    const renderColumn = (title: string, tasks: Task[], droppableId: string) => (
        <div className="card-list">
            <h2>{title}</h2>
            <SortableContext id={droppableId} items={tasks.map(task => task.id)} strategy={verticalListSortingStrategy}>
                <div className="task-cards">
                    {tasks.length === 0 && (<PlaceHolder key={'empty'} />)}
                    {tasks.map(task => <SortableItem key={task.id} task={task} />)}
                    {droppableId === 'Planned' && showInputPlanned && (
                        <>
                            <div style={{ borderTop: "2px solid black", marginLeft: 20, marginRight: 20 }}></div>
                            <div className="add-card-form">
                                <input
                                    type="text"
                                    name="title"
                                    placeholder="Title (must be filled)"
                                    value={newTask.title}
                                    onChange={handleNewTaskChange}
                                />
                                <textarea
                                    name="description"
                                    placeholder="Description"
                                    value={newTask.description}
                                    onChange={handleNewTaskChange}
                                />
                                <FibonacciSelector
                                    selectedValue={parseInt(newTask.estimated_time)}
                                    onSelect={handleFibonacciSelect}
                                />
                                <input
                                    type="number"
                                    name="actual_time"
                                    placeholder="Actual Time"
                                    value={newTask.actual_time}
                                    onChange={handleNewTaskChange}
                                />
                                <input
                                    type="date"
                                    name="due_date"
                                    placeholder="Due Date (must be filled)"
                                    value={newTask.due_date}
                                    onChange={handleNewTaskChange}
                                />
                                <input
                                    type="number"
                                    name="priority"
                                    placeholder="Priority"
                                    value={newTask.priority}
                                    onChange={handleNewTaskChange}
                                />
                                <input
                                    type="number"
                                    name="assignee_id"
                                    placeholder="Assignee ID (must be filled)"
                                    value={newTask.assignee_id}
                                    onChange={handleNewTaskChange}
                                />
                                <input
                                    type="text"
                                    name="image_url"
                                    placeholder="Image URL"
                                    value={newTask.image_url}
                                    onChange={handleNewTaskChange}
                                />
                                <h3>Upload a Picture</h3>
                                <input
                                    type="file"
                                    name="image"
                                    onChange={onFileChange}
                                />
                                <button onClick={() => handleAddCard('Planned')}>Add Card</button>
                                <button className="cancel-btn" onClick={() => handleCancelAddCard('Planned')}>
                                    Cancel
                                </button>
                            </div>
                        </>
                    )}
                    {!showInputPlanned && droppableId === 'Planned' && (
                        <button className="add-card-btn" onClick={() => handleAddCardClick('Planned')}>
                            + Add a card
                        </button>
                    )}
                </div>
            </SortableContext>
        </div>
    );

    const sensors = useSensors(
        useSensor(PointerSensor, {
            activationConstraint: {
                distance: 8,
            },
        }),
        useSensor(KeyboardSensor, {
            coordinateGetter: sortableKeyboardCoordinates,
        })
    );

    return (
        <div className="workspace-page">
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
                    <button className="back-btn" onClick={handleExitWorkspace}>
                        Back
                    </button>
                </div>
                <div className="right-container">
                    <button className="leave-workspace-btn" onClick={handleLeaveWorkspace}>
                        Leave Workspace
                    </button>
                    <button className="delete-btn" onClick={handleDeleteWorkspace}>
                        Delete Workspace
                    </button>
                </div>
            </header>
            <div className="main-content">
                <aside className="members-section">
                    {membersUtils.map(member_utils => {
                        return (
                            <div
                                className="member"
                                onClick={() => handleMemberClick(member_utils)}
                            >
                                {member_utils.member_profile.status == "online" && (
                                    <div className="status-icon">
                                        <img src={online_icon} />
                                        {member_utils.member_profile.username}
                                    </div>
                                )}
                                {member_utils.member_profile.status == "offline" && (
                                    <div className="status-icon">
                                        <img src={offline_icon} />
                                        {member_utils.member_profile.username}
                                    </div>
                                )}
                            </div>
                        )
                    })}
                </aside>
                <div className="tasks-section">
                    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleOnDragEnd}>
                        <div className="cards-section">
                            {renderColumn('Planned', taskPlanned, 'Planned')}
                            {renderColumn('InProgress', taskInProgress, 'InProgress')}
                            {renderColumn('Completed', taskCompleted, 'Completed')}
                        </div>
                    </DndContext>
                </div>
            </div >
            {
                isTaskPageVisible && selectedTask && (
                    <TaskPage
                        task={selectedTask}
                        workspace_id={workspace_id!}
                        onClose={() => {
                            setTaskPageVisible(false);
                            Websocket.close();
                            Websocket.connect(websocketMessageHandler);
                            fetchTasks();
                            fetchMembers();
                        }}
                        onTaskUpdate={fetchTasks}
                    />
                )
            }
            {
                isMemberDetailsVisible && selectedMemberUtils && (
                    <MemberDetails
                        member_utils={selectedMemberUtils}
                        onClose={() => {
                            setMemberDetailsVisible(false);
                            Websocket.close();
                            Websocket.connect(websocketMessageHandler);
                            fetchTasks();
                            fetchMembers();
                        }}
                    />
                )
            }
        </div >
    );
};

export default Workspace;
