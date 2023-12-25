local function start()
    box.schema.space.create("sessions")
    box.space.sessions:format({ { name = "uuid", type = "string", is_nullable = false }, { name = "expire_time", type = "datetime", is_nullable = false }, { name = "email", type = "string", is_nullable = false } })
    box.space.sessions:create_index("primary", { parts = { "uuid" } })
end

return {
    start = start
}