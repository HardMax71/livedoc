import { Component, For } from 'solid-js';
import { collaborationStore } from '@/stores/collaboration.ts';
import { format } from 'date-fns';

const Collaboration: Component = () => {
    return (
        <div class="collaboration-panel">
            <h3 class="title is-5">Active Users</h3>
            <div class="active-users">
                <For each={collaborationStore.state.activeUsers}>
                    {(user) => (
                        <div class="active-user">
                            <span class="user-indicator"></span>
                            <div class="user-info">
                                <span class="username">{user.username}</span>
                                <span class="last-active">
                  {format(new Date(user.lastActive), 'HH:mm:ss')}
                </span>
                            </div>
                        </div>
                    )}
                </For>
            </div>
        </div>
    );
};

export default Collaboration;