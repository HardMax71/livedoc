import { Component, createEffect, For } from 'solid-js';
import { useNavigate } from '@solidjs/router';
import Button from '../../components/common/Button';
import { documentStore } from '@/stores/document.ts';
import { format } from 'date-fns';

const DocumentList: Component = () => {
    const navigate = useNavigate();

    createEffect(() => {
        documentStore.listDocuments();
    });

    return (
        <div class="container">
            <div class="level mt-4">
                <div class="level-left">
                    <h1 class="title">My Documents</h1>
                </div>
                <div class="level-right">
                    <Button onClick={() => navigate('/documents/new')}>
                        New Document
                    </Button>
                </div>
            </div>

            <div class="documents-grid">
                <For each={documentStore.state.documents}>
                    {(doc) => (
                        <div class="box document-card">
                            <h3 class="title is-5">{doc.title}</h3>
                            <p class="subtitle is-7">
                                Last edited: {format(new Date(doc.updatedAt), 'MMM dd, yyyy HH:mm')}
                            </p>
                            <div class="buttons are-small">
                                <Button onClick={() => navigate(`/documents/${doc.id}`)}>
                                    Open
                                </Button>
                                <Button
                                    variant="danger"
                                    onClick={() => documentStore.deleteDocument(doc.id)}
                                >
                                    Delete
                                </Button>
                            </div>
                        </div>
                    )}
                </For>
            </div>

            {documentStore.state.error && (
                <p class="has-text-danger">{documentStore.state.error}</p>
            )}
        </div>
    );
};

export default DocumentList;