if ! docker exec ceph-aio radosgw-admin user info --uid=panyu >/dev/null 2>&1; then
    echo "创建 panyu 用户"
    docker exec ceph-aio radosgw-admin user create \
        --uid=panyu \
        --display-name="Panyu" \
        --access-key=panyu \
        --secret-key=panyu
fi
