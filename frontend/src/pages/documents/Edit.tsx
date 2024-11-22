import { Component } from 'solid-js';
import { useParams } from '@solidjs/router';
import Editor from '../../components/editor/Editor';
import { documentStore } from '@/stores/document.ts';
import Loading from '../../components/common/Loading';

const EditDocument: Component = () => {
    const params = useParams();

    return (
        <div class="container is-fluid">
            {documentStore.state.isLoading ? (
                <Loading />
            ) : documentStore.state.currentDocument ? (
                <Editor documentId={params.id} />
            ) : (
                <div class="has-text-centered mt-6">
                    <p class="has-text-danger">Document not found</p>
                </div>
            )}
        </div>
    );
};

export default EditDocument;