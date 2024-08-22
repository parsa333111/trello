import { FormEvent, useState } from "react";
import '../styles/MemberDetails.css';
import { MessageType, toastifyMessage } from "./Toastify";
import axios from "axios";

export interface Member {
    id: string;
    user_id: string;
    workspace_id: string;
    role: string;
    created_at: string;
    updated_at: string;
}

export interface MemberProfile {
    id: string;
    username: string;
    email: string;
    created_at: string;
    status: string;
}

export interface MemberUtils {
    member: Member;
    member_profile: MemberProfile;
}

interface MemberDetailsProps {
    member_utils: MemberUtils;
    onClose: () => void;
}

const MemberDetails: React.FC<MemberDetailsProps> = ({ member_utils, onClose }) => {

    const [newRole, setNewRole] = useState("");

    const handleChangeRole = async (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();

        try {
            await axios.put(`/api/workspaces/${member_utils.member.workspace_id}/users/${member_utils.member.user_id}`, {
                role: newRole,
            });
            toastifyMessage(`Role has been changed successfully.`, MessageType.Success);
            onClose();
        } catch (error) {
            toastifyMessage('Failed to change role.', MessageType.Error);
        }

        setNewRole("");
    };

    const handleKickMember = async () => {
        const confirmed = window.confirm("Are you sure you want to kick out  this member? This action cannot be undone.");

        if (confirmed) {
            try {
                await axios.delete(`/api/workspaces/${member_utils.member.workspace_id}/users/${member_utils.member.user_id}`);
                toastifyMessage(`Member has been kicked out successfully.`, MessageType.Success);
                onClose();
            } catch (error) {
                toastifyMessage('Failed to kick member.', MessageType.Error);
            }
        }

        setNewRole("");
    };

    return (
        <div className="member-details-modal">
            <div className="member-details-content">
                <div className="member-details-info">
                    <h2>Member Details</h2>
                    <p><strong>User ID:</strong> {member_utils.member.user_id}</p>
                    <p><strong>Username:</strong> {member_utils.member_profile.username}</p>
                    <p><strong>Email:</strong> {member_utils.member_profile.email}</p>
                    <p><strong>Status:</strong> {member_utils.member_profile.status}</p>
                    <p><strong>Role:</strong> {member_utils.member.role}</p>
                    <button type="button" className='kick-button' onClick={handleKickMember}>Kick Out</button>
                    <div className="change-role-form">
                        <form onSubmit={(event) => handleChangeRole(event)}>
                            <label>Change Role</label>
                            <input
                                type="text"
                                value={newRole}
                                onChange={(event) => setNewRole(event.target.value)}
                            />
                            <div className="button-group">
                                <button type="submit" className='change-button'>Change</button>
                                <button type="button" className="close-button" onClick={onClose}>Close</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default MemberDetails;
