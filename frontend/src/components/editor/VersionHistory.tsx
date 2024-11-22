import { Component, createEffect, For } from 'solid-js';
import { documentStore } from '@/stores/document.ts';
import Button from '../common/Button';
import { format } from 'date-fns';

interface VersionHistoryProps {
    documentId: string;
}

const VersionHistory: Component<VersionHistoryProps> = (props) => {
    createEffect(() => {
        documentStore.getVersionHistory(props.documentId);
    });

    const handleRestore = async (versionId: string) => {
        await documentStore.restoreVersion(props.documentId, versionId);
    };

    return (
        <div class="version-history-panel">
            <h3 class="title is-5">Version History</h3>
            <div class="versions">
                <For each={documentStore.state.versions}>
                    {(version) => (
                        <div class="version-item">
                            <div class="version-info">
                <span class="timestamp">
                  {format(new Date(version.createdAt), 'MMM dd, HH:mm')}
                </span>
                                <Button
                                    variant="ghost"
                                    size="small"
                                    onClick={() => handleRestore(version.id)}
                                >
                                    Restore
                                </Button>
                            </div>
                        </div>
                    )}
                </For>
            </div>
        </div>
    );
};

export default VersionHistory;