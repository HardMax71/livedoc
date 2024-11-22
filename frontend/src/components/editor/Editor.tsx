import { Component, createEffect, createSignal, onCleanup } from 'solid-js';
import { documentStore } from '../../stores/document';
import { collaborationStore } from '../../stores/collaboration';
import { Operation, OperationType } from '../../types/collaboration';
import Toolbar from './Toolbar';
import Collaboration from './Collaboration';
import VersionHistory from './VersionHistory';

interface EditorProps {
    documentId: string;
}

const Editor: Component<EditorProps> = (props) => {
    let editorRef: HTMLDivElement | undefined;
    const [content, setContent] = createSignal('');
    const [version, setVersion] = createSignal('');
    const [isSyncing, setIsSyncing] = createSignal(false);

    createEffect(async () => {
        await documentStore.getDocument(props.documentId);
        if (documentStore.state.currentDocument) {
            setContent(documentStore.state.currentDocument.content);
            setVersion(documentStore.state.currentDocument.version);
        }
    });

    createEffect(async () => {
        await collaborationStore.joinSession(props.documentId);

        collaborationStore.onDocumentChange(props.documentId, (change) => {
            if (change.userId !== collaborationStore.state.sessionId) {
                applyChanges(change.operations);
                setVersion(change.version);
            }
        });
    });

    onCleanup(async () => {
        await collaborationStore.leaveSession();
    });

    const applyChanges = (operations: Operation[]) => {
        let newContent = content();
        for (const op of operations) {
            switch (op.type) {
                case OperationType.INSERT:
                    newContent = newContent.slice(0, op.position) +
                        op.content +
                        newContent.slice(op.position);
                    break;
                case OperationType.DELETE:
                    newContent = newContent.slice(0, op.position) +
                        newContent.slice(op.position + op.length);
                    break;
                case OperationType.REPLACE:
                    newContent = newContent.slice(0, op.position) +
                        op.content +
                        newContent.slice(op.position + op.length);
                    break;
            }
        }
        setContent(newContent);
    };

    const handleInput = async (event: InputEvent) => {
        const target = event.target as HTMLDivElement;
        const newContent = target.innerText;

        if (!isSyncing()) {
            setIsSyncing(true);

            const operations: Operation[] = [];
            // Simple diff algorithm - in real implementation, use a proper diff library
            if (newContent.length > content().length) {
                operations.push({
                    type: OperationType.INSERT,
                    position: content().length,
                    content: newContent.slice(content().length),
                    length: 0,
                });
            } else {
                operations.push({
                    type: OperationType.DELETE,
                    position: newContent.length,
                    content: '',
                    length: content().length - newContent.length,
                });
            }

            const response = await collaborationStore.syncDocument(
                props.documentId,
                operations,
                version()
            );

            if (response) {
                setVersion(response.newVersion);
                if (response.concurrentChanges.length > 0) {
                    response.concurrentChanges.forEach(change => {
                        applyChanges(change.operations);
                    });
                }
            }

            setIsSyncing(false);
        }

        setContent(newContent);
    };

    return (
        <div class="editor-container">
            <Toolbar documentId={props.documentId} />
            <div class="columns">
                <div class="column is-9">
                    <div
                        ref={editorRef}
                        class="editor-content"
                        contentEditable={true}
                        onInput={handleInput}
                        innerHTML={content()}
                    />
                </div>
                <div class="column is-3">
                    <Collaboration />
                    <VersionHistory documentId={props.documentId} />
                </div>
            </div>
        </div>
    );
};

export default Editor;